package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"sync"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/spv"
	"github.com/bsv-blockchain/go-sdk/transaction"

	"github.com/bsv-blockchain/go-overlay-services/pkg/core/gasp"
)

var (
	// ErrGraphFull indicates the graph has reached its maximum size
	ErrGraphFull = errors.New("graph is full")

	// ErrParsedBEEFReturnedNilTx indicates that parsing BEEF returned a nil transaction
	ErrParsedBEEFReturnedNilTx = errors.New("parsed BEEF returned nil transaction")

	// ErrGraphAnchorInvalidTx indicates that the graph anchor is not a valid transaction
	ErrGraphAnchorInvalidTx = errors.New("graph anchor is not a valid transaction")

	// ErrGraphNoTopicalAdmittance indicates that the graph did not result in topical admittance of the root node
	ErrGraphNoTopicalAdmittance = errors.New("graph did not result in topical admittance of the root node. rejecting")
	// ErrUnableToFindRootNodeInGraph indicates that the root node could not be found in the graph for finalization
	ErrUnableToFindRootNodeInGraph = errors.New("unable to find root node in graph for finalization")
	// ErrRequiredInputNodeNotFoundInTempGraph indicates that a required input node was not found in the temporary graph store
	ErrRequiredInputNodeNotFoundInTempGraph = errors.New("required input node for unproven parent not found in temporary graph store")
)

// submissionState tracks the state of a transaction submission
type submissionState struct {
	wg  sync.WaitGroup
	err error
}

// GraphNode represents a node in the GASP graph
type GraphNode struct {
	gasp.Node

	Txid     *chainhash.Hash `json:"txid"`
	SpentBy  *chainhash.Hash `json:"spentBy"`
	Children sync.Map        `json:"-"` // map[string]*GraphNode - concurrent safe
	Parent   *GraphNode      `json:"parent"`
}

// OverlayGASPStorage implements GASP storage using the overlay engine
type OverlayGASPStorage struct {
	Topic              string
	Engine             *Engine
	MaxNodesInGraph    *int
	tempGraphNodeRefs  sync.Map
	tempGraphNodeCount int
	submissionTracker  sync.Map // map[chainhash.Hash]*submissionState
}

// NewOverlayGASPStorage creates a new OverlayGASPStorage instance
func NewOverlayGASPStorage(topic string, engine *Engine, maxNodesInGraph *int) *OverlayGASPStorage {
	return &OverlayGASPStorage{
		Topic:           topic,
		Engine:          engine,
		MaxNodesInGraph: maxNodesInGraph,
	}
}

// ErrNoKnownUTXOs is returned when no UTXOs are found
var ErrNoKnownUTXOs = errors.New("no known UTXOs")

// FindKnownUTXOs retrieves known UTXOs for the topic
func (s *OverlayGASPStorage) FindKnownUTXOs(ctx context.Context, since float64, limit uint32) ([]*gasp.Output, error) {
	utxos, err := s.Engine.Storage.FindUTXOsForTopic(ctx, s.Topic, since, limit, false)
	if err != nil {
		return nil, err
	}
	gaspOutputs := make([]*gasp.Output, len(utxos))

	for i, utxo := range utxos {
		gaspOutputs[i] = &gasp.Output{
			Txid:        utxo.Outpoint.Txid,
			OutputIndex: utxo.Outpoint.Index,
			Score:       utxo.Score,
		}
	}

	return gaspOutputs, nil
}

func (s *OverlayGASPStorage) HasOutputs(ctx context.Context, outpoints []*transaction.Outpoint) ([]bool, error) {
	// Use FindOutputs to check existence - don't need BEEF for existence check
	outputs, err := s.Engine.Storage.FindOutputs(ctx, outpoints, s.Topic, nil, false)
	if err != nil {
		return nil, err
	}

	// Convert to boolean array - true if output exists, false if nil
	result := make([]bool, len(outputs))
	for i, output := range outputs {
		result[i] = output != nil
	}
	return result, nil
}

// HydrateGASPNode hydrates a GASP node from storage
func (s *OverlayGASPStorage) HydrateGASPNode(ctx context.Context, graphID, outpoint *transaction.Outpoint, _ bool) (*gasp.Node, error) {
	output, err := s.Engine.Storage.FindOutput(ctx, outpoint, nil, nil, true)
	if err != nil {
		return nil, err
	}
	if output == nil || output.Beef == nil {
		return nil, ErrMissingInput
	}
	// Parse BEEF to get the transaction
	_, tx, _, err := transaction.ParseBeef(output.Beef)
	if err != nil {
		return nil, err
	}

	// Check if we got a valid transaction
	if tx == nil {
		return nil, ErrParsedBEEFReturnedNilTx
	}

	node := &gasp.Node{
		GraphID:     graphID,
		OutputIndex: outpoint.Index,
		RawTx:       tx.Hex(),
	}
	if tx.MerklePath != nil {
		proof := tx.MerklePath.Hex()
		node.Proof = &proof
	}
	return node, nil
}

// ErrNoNeededInputs is returned when no inputs are needed
var ErrNoNeededInputs = errors.New("no needed inputs")

// FindNeededInputs determines which inputs are needed for a GASP transaction
func (s *OverlayGASPStorage) FindNeededInputs(ctx context.Context, gaspTx *gasp.Node) (*gasp.NodeResponse, error) {
	response := &gasp.NodeResponse{
		RequestedInputs: make(map[transaction.Outpoint]*gasp.NodeResponseData),
	}
	tx, err := transaction.NewTransactionFromHex(gaspTx.RawTx)
	if err != nil {
		return nil, err
	}
	// Commented out: This was requesting ALL inputs for unmined transactions
	// but should use IdentifyNeededInputs to get only relevant inputs
	if gaspTx.Proof == nil || *gaspTx.Proof == "" {
		for _, input := range tx.Inputs {
			outpoint := &transaction.Outpoint{
				Txid:  *input.SourceTXID,
				Index: input.SourceTxOutIndex,
			}
			response.RequestedInputs[*outpoint] = &gasp.NodeResponseData{
				Metadata: false,
			}
		}

		return s.stripAlreadyKnowInputs(ctx, response)
	}

	// Process merkle proof if present
	if gaspTx.Proof != nil && *gaspTx.Proof != "" {
		if tx.MerklePath, err = transaction.NewMerklePathFromHex(*gaspTx.Proof); err != nil {
			return nil, err
		}
	}

	var beef *transaction.Beef
	if tx.MerklePath != nil {
		// If we have a merkle path, create BEEF from transaction
		if beef, err = transaction.NewBeefFromTransaction(tx); err != nil {
			return nil, err
		}
	}

	if beef != nil {
		inpoints := make([]*transaction.Outpoint, len(tx.Inputs))
		for vin, input := range tx.Inputs {
			inpoints[vin] = &transaction.Outpoint{
				Txid:  *input.SourceTXID,
				Index: input.SourceTxOutIndex,
			}
		}
		previousCoins := make([]uint32, 0, len(tx.Inputs))
		outputs, err := s.Engine.Storage.FindOutputs(ctx, inpoints, s.Topic, nil, true)
		if err != nil {
			return nil, err
		}
		for vin, output := range outputs {
			if output != nil {
				beef.MergeBeefBytes(output.Beef)
				previousCoins = append(previousCoins, uint32(vin))
			}
		}

		beefBytes, err := beef.AtomicBytes(tx.TxID())
		if err != nil {
			return nil, err
		}
		admit, _ := s.IdentifyAdmissibleOutputs(ctx, beefBytes, previousCoins)
		if !slices.Contains(admit.OutputsToAdmit, gaspTx.OutputIndex) {
			neededInputs, err := s.IdentifyNeededInputs(ctx, beefBytes)
			if err != nil {
				return nil, err
			}
			for _, outpoint := range neededInputs {
				response.RequestedInputs[*outpoint] = &gasp.NodeResponseData{
					Metadata: true,
				}
			}
			return s.stripAlreadyKnowInputs(ctx, response)
		}
	}

	return response, nil
}

func (s *OverlayGASPStorage) IdentifyAdmissibleOutputs(ctx context.Context, beefBytes []byte, previousCoins []uint32) (overlay.AdmittanceInstructions, error) {
	if _, ok := s.Engine.Managers[s.Topic]; !ok {
		return overlay.AdmittanceInstructions{}, errors.New("no manager for topic (identify admissible outputs): " + s.Topic)
	}
	return s.Engine.Managers[s.Topic].IdentifyAdmissibleOutputs(ctx, beefBytes, previousCoins)
}

func (s *OverlayGASPStorage) IdentifyNeededInputs(ctx context.Context, beefBytes []byte) ([]*transaction.Outpoint, error) {
	if _, ok := s.Engine.Managers[s.Topic]; !ok {
		return nil, errors.New("no manager for topic (identify needed inputs): " + s.Topic)
	}
	return s.Engine.Managers[s.Topic].IdentifyNeededInputs(ctx, beefBytes)
}

// ErrNoInputsToStrip is returned when there are no inputs to strip
var ErrNoInputsToStrip = errors.New("no inputs to strip")

func (s *OverlayGASPStorage) stripAlreadyKnowInputs(ctx context.Context, response *gasp.NodeResponse) (*gasp.NodeResponse, error) {
	if response == nil {
		return nil, ErrNoInputsToStrip
	}
	for outpoint := range response.RequestedInputs {
		if found, err := s.Engine.Storage.FindOutput(ctx, &outpoint, &s.Topic, nil, false); err != nil {
			return nil, err
		} else if found != nil {
			delete(response.RequestedInputs, outpoint)
		}
	}
	if len(response.RequestedInputs) == 0 {
		return nil, ErrNoInputsToStrip
	}
	return response, nil
}

// AppendToGraph adds a GASP node to the temporary graph store for later validation and finalization.
func (s *OverlayGASPStorage) AppendToGraph(_ context.Context, gaspTx *gasp.Node, spentBy *transaction.Outpoint) error {
	if s.MaxNodesInGraph != nil && s.tempGraphNodeCount >= *s.MaxNodesInGraph {
		return ErrGraphFull
	}

	tx, err := transaction.NewTransactionFromHex(gaspTx.RawTx)
	if err != nil {
		return err
	}
	txid := tx.TxID()
	if gaspTx.Proof != nil && *gaspTx.Proof != "" {
		if tx.MerklePath, err = transaction.NewMerklePathFromHex(*gaspTx.Proof); err != nil {
			slog.Error("Failed to parse merkle path", "error", err, "proofLength", len(*gaspTx.Proof))
			return err
		}
	}
	newGraphNode := &GraphNode{
		Node: *gaspTx,
		Txid: txid,
	}
	// Compute the actual outpoint from the returned transaction
	newGraphOutpoint := &transaction.Outpoint{
		Txid:  *txid,
		Index: gaspTx.OutputIndex,
	}

	// Store the node by its actual outpoint (not by GraphID)
	if _, ok := s.tempGraphNodeRefs.LoadOrStore(*newGraphOutpoint, newGraphNode); !ok {
		s.tempGraphNodeCount++
	}

	// If this node has a parent, link them together
	if spentBy != nil {
		if parentNode, ok := s.tempGraphNodeRefs.Load(*spentBy); !ok {
			return ErrMissingInput
		} else {
			parent := parentNode.(*GraphNode)
			parent.Children.Store(*newGraphOutpoint, newGraphNode)
			newGraphNode.Parent = parentNode.(*GraphNode)
		}
	}
	return nil
}

// ValidateGraphAnchor verifies that the graph anchor transaction is valid and results in topical admittance.
func (s *OverlayGASPStorage) ValidateGraphAnchor(ctx context.Context, graphID *transaction.Outpoint) error {
	if rootNode, ok := s.tempGraphNodeRefs.Load(*graphID); !ok {
		return ErrMissingInput
	} else if beef, err := s.getBEEFForNode(rootNode.(*GraphNode)); err != nil {
		return err
	} else if tx, err := transaction.NewTransactionFromBEEF(beef); err != nil {
		return err
	} else if valid, err := spv.Verify(ctx, tx, s.Engine.ChainTracker, nil); err != nil {
		return err
	} else if !valid {
		return ErrGraphAnchorInvalidTx
	}
	beefs, beefsErr := s.computeOrderedBEEFsForGraph(ctx, graphID)
	if beefsErr != nil {
		return beefsErr
	}
	coins := make(map[transaction.Outpoint]struct{})
	for _, beefBytes := range beefs {
		if beef, tx, txid, err := transaction.ParseBeef(beefBytes); err != nil {
			return err
		} else {
			inpoints := make([]*transaction.Outpoint, len(tx.Inputs))
			for vin, input := range tx.Inputs {
				inpoints[vin] = &transaction.Outpoint{
					Txid:  *input.SourceTXID,
					Index: input.SourceTxOutIndex,
				}
			}
			previousCoins := make([]uint32, 0, len(tx.Inputs))
			if outputs, err := s.Engine.Storage.FindOutputs(ctx, inpoints, s.Topic, nil, true); err != nil {
				return err
			} else {
				for vin, output := range outputs {
					if output != nil {
						beef.MergeBeefBytes(output.Beef)
						previousCoins = append(previousCoins, uint32(vin))
					}
				}
			}
			if beefBytes, err = beef.AtomicBytes(txid); err != nil {
				return err
			} else if admit, err := s.IdentifyAdmissibleOutputs(ctx, beefBytes, previousCoins); err != nil {
				slog.Error("[GASP] ValidateGraphAnchor failed to identify admissible outputs", "error", err)
				return err
			} else {
				for _, vout := range admit.OutputsToAdmit {
					outpoint := &transaction.Outpoint{
						Txid:  *tx.TxID(),
						Index: vout,
					}
					coins[*outpoint] = struct{}{}
				}
			}
		}
	}
	if _, ok := coins[*graphID]; !ok {
		return ErrGraphNoTopicalAdmittance
	}
	return nil
}

// DiscardGraph removes all nodes associated with the specified graph from the temporary storage.
func (s *OverlayGASPStorage) DiscardGraph(_ context.Context, graphID *transaction.Outpoint) error {
	// Find and delete all nodes that belong to this graph
	nodesToDelete := make([]*transaction.Outpoint, 0)

	// First pass: collect all node IDs that belong to this graph
	s.tempGraphNodeRefs.Range(func(nodeId, graphRef any) bool {
		node := graphRef.(*GraphNode)
		if node.GraphID.Equal(graphID) {
			outpoint := nodeId.(transaction.Outpoint)
			nodesToDelete = append(nodesToDelete, &outpoint)
		}
		return true
	})

	// Delete all collected nodes
	for _, nodeID := range nodesToDelete {
		s.tempGraphNodeRefs.Delete(*nodeID)
		s.tempGraphNodeCount--
	}

	return nil
}

// FinalizeGraph submits all transactions in the graph to the overlay engine for processing.
func (s *OverlayGASPStorage) FinalizeGraph(ctx context.Context, graphID *transaction.Outpoint) error {
	beefs, err := s.computeOrderedBEEFsForGraph(ctx, graphID)
	if err != nil {
		return err
	}
	for _, beef := range beefs {
		// Extract transaction ID from BEEF for deduplication key
		_, tx, txid, err := transaction.ParseBeef(beef)
		if err != nil {
			return err
		}
		if tx == nil {
			return errors.New("no transaction in BEEF")
		}

		// Pre-initialize the submission state to avoid race conditions
		newState := &submissionState{}
		newState.wg.Add(1)

		if existing, loaded := s.submissionTracker.LoadOrStore(txid, newState); loaded {
			// Another goroutine is already submitting this transaction, wait for it
			state := existing.(*submissionState)
			state.wg.Wait()
			if state.err != nil {
				return state.err
			}
		} else {
			// We're the first caller, do the submission using our pre-initialized state
			state := newState
			defer state.wg.Done() // Signal completion

			// Perform the actual submission
			_, state.err = s.Engine.Submit(
				ctx,
				overlay.TaggedBEEF{
					Topics: []string{s.Topic},
					Beef:   beef,
				},
				SubmitModeHistorical,
				nil,
			)
			if state.err != nil {
				slog.Error("[GASP] Failed to submit transaction", "txid", txid.String(), "error", state.err)
				return state.err
			}
			slog.Info(fmt.Sprintf("[GASP] Transaction processed: %s", txid.String()))
		}
	}
	return nil
}

func (s *OverlayGASPStorage) computeOrderedBEEFsForGraph(_ context.Context, graphID *transaction.Outpoint) ([][]byte, error) {
	beefs := make([][]byte, 0)
	var hydrator func(node *GraphNode) error
	hydrator = func(node *GraphNode) error {
		currentBeef, err := s.getBEEFForNode(node)
		if err != nil {
			return err
		}
		if slices.IndexFunc(beefs, func(beef []byte) bool {
			return bytes.Equal(beef, currentBeef)
		}) == -1 {
			beefs = append([][]byte{currentBeef}, beefs...)
		}
		var childErr error
		node.Children.Range(func(key, value any) bool {
			child := value.(*GraphNode)
			if err := hydrator(child); err != nil {
				childErr = err
				return false
			}
			return true
		})
		if childErr != nil {
			return childErr
		}
		return nil
	}

	foundRoot, ok := s.tempGraphNodeRefs.Load(*graphID)
	if !ok {
		return nil, ErrUnableToFindRootNodeInGraph
	}
	if err := hydrator(foundRoot.(*GraphNode)); err != nil {
		return nil, err
	}
	return beefs, nil
}

func (s *OverlayGASPStorage) getBEEFForNode(node *GraphNode) ([]byte, error) {
	if node == nil {
		panic(fmt.Sprintf("GASP DEBUG: getBEEFForNode called with nil node. Total goroutines: %d", runtime.NumGoroutine()))
	}

	var hydrator func(node *GraphNode) (*transaction.Transaction, error)
	hydrator = func(node *GraphNode) (*transaction.Transaction, error) {
		if node == nil {
			panic(fmt.Sprintf("GASP DEBUG: hydrator called with nil node. Goroutine: %d", runtime.NumGoroutine()))
		}
		tx, err := transaction.NewTransactionFromHex(node.RawTx)
		if err != nil {
			return nil, err
		}
		if node.Proof != nil && *node.Proof != "" {
			if tx.MerklePath, err = transaction.NewMerklePathFromHex(*node.Proof); err != nil {
				return nil, err
			}
			return tx, nil
		}
		for vin, input := range tx.Inputs {
			outpoint := &transaction.Outpoint{
				Txid:  *input.SourceTXID,
				Index: input.SourceTxOutIndex,
			}
			foundNode, ok := s.tempGraphNodeRefs.Load(*outpoint)
			if !ok {
				return nil, ErrRequiredInputNodeNotFoundInTempGraph
			}
			if tx.Inputs[vin].SourceTransaction, err = hydrator(foundNode.(*GraphNode)); err != nil {
				return nil, err
			}
		}
		return tx, nil
	}
	tx, err := hydrator(node)
	if err != nil {
		return nil, err
	}
	beef, err := transaction.NewBeefFromTransaction(tx)
	if err != nil {
		return nil, err
	}
	return beef.AtomicBytes(tx.TxID())
}
