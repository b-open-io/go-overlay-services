package gasp

import (
	"bytes"
	"context"
	"errors"
	"slices"

	"github.com/4chain-ag/go-overlay-services/pkg/core/engine"
	"github.com/4chain-ag/go-overlay-services/pkg/core/gasp/core"
	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/spv"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

var ErrGraphFull = errors.New("graph is full")

type overlayGASPStorage struct {
	Topic             string
	Engine            engine.Engine
	MaxNodesInGraph   *int
	tempGraphNodeRefs map[string]*GraphNode
}

func NewOverlayGASPStorage(topic string, engine engine.Engine, maxNodesInGraph *int) *overlayGASPStorage {
	return &overlayGASPStorage{
		Topic:             topic,
		Engine:            engine,
		MaxNodesInGraph:   maxNodesInGraph,
		tempGraphNodeRefs: make(map[string]*GraphNode),
	}
}

func (s *overlayGASPStorage) FindKnownUTXOs(ctx context.Context, since uint32) ([]*overlay.Outpoint, error) {
	if utxos, err := s.Engine.Storage.FindUTXOsForTopic(ctx, s.Topic, since, false); err != nil {
		return nil, err
	} else {
		outpoints := make([]*overlay.Outpoint, len(utxos))
		for i, utxo := range utxos {
			outpoints[i] = &utxo.Outpoint
		}

		return outpoints, nil
	}
}

func (s *overlayGASPStorage) HydrateGASPNode(ctx context.Context, graphID *overlay.Outpoint, outpoint *overlay.Outpoint, metadata bool) (*core.GASPNode, error) {
	if output, err := s.Engine.Storage.FindOutput(ctx, outpoint, nil, nil, true); err != nil {
		return nil, err
	} else if output.Beef == nil {
		return nil, engine.ErrMissingInput
	} else if tx, err := transaction.NewTransactionFromBEEF(output.Beef); err != nil {
		return nil, err
	} else {
		node := &core.GASPNode{
			GraphID:     graphID,
			OutputIndex: outpoint.OutputIndex,
			RawTx:       tx.Hex(),
		}
		if tx.MerklePath != nil {
			node.Proof = tx.MerklePath.Hex()
		}
		return node, nil
	}
}

func (s *overlayGASPStorage) FindNeededInputs(ctx context.Context, gaspTx *core.GASPNode) (*core.GASPNodeResponse, error) {
	response := &core.GASPNodeResponse{
		RequestedInputs: make(map[string]*core.GASPNodeResponseData),
	}

	if tx, err := transaction.NewTransactionFromHex(gaspTx.RawTx); err != nil {
		return nil, err
	} else {
		for _, input := range tx.Inputs {
			outpoint := &overlay.Outpoint{
				Txid:        *input.SourceTXID,
				OutputIndex: input.SourceTxOutIndex,
			}
			response.RequestedInputs[outpoint.String()] = &core.GASPNodeResponseData{
				Metadata: false,
			}
		}

		return s.stripAlreadyKnowInputs(ctx, response)
	}
}

func (s *overlayGASPStorage) stripAlreadyKnowInputs(ctx context.Context, response *core.GASPNodeResponse) (*core.GASPNodeResponse, error) {
	if response == nil {
		return nil, nil
	}
	for outpointStr := range response.RequestedInputs {
		if outpoint, err := overlay.NewOutpointFromString(outpointStr); err != nil {
			return nil, err
		} else if found, err := s.Engine.Storage.FindOutput(ctx, outpoint, &s.Topic, nil, false); err != nil {
			return nil, err
		} else if found != nil {
			delete(response.RequestedInputs, outpointStr)
		}
	}
	if len(response.RequestedInputs) == 0 {
		return nil, nil
	}
	return response, nil
}

func (s *overlayGASPStorage) AppendToGraph(ctx context.Context, gaspTx *core.GASPNode, spentBy *chainhash.Hash) error {
	if s.MaxNodesInGraph != nil && len(s.tempGraphNodeRefs) >= *s.MaxNodesInGraph {
		return ErrGraphFull
	}

	if tx, err := transaction.NewTransactionFromHex(gaspTx.RawTx); err != nil {
		return err
	} else {
		txid := tx.TxID()
		if len(gaspTx.Proof) > 0 {
			if tx.MerklePath, err = transaction.NewMerklePathFromHex(gaspTx.Proof); err != nil {
				return err
			}
		}
		newGraphNode := &GraphNode{
			Txid:           txid,
			GraphID:        gaspTx.GraphID,
			RawTx:          gaspTx.RawTx,
			OutputIndex:    gaspTx.OutputIndex,
			Proof:          gaspTx.Proof,
			TxMetadata:     gaspTx.TxMetadata,
			OutputMetadata: gaspTx.OutputMetadata,
			Inputs:         gaspTx.Inputs,
			Children:       []*GraphNode{},
		}
		if spentBy != nil {
			s.tempGraphNodeRefs[spentBy.String()] = newGraphNode
		} else if parentNode, ok := s.tempGraphNodeRefs[txid.String()]; ok {
			parentNode.Children = append(parentNode.Children, newGraphNode)
			newGraphNode.Parent = parentNode
			newGraphOutpoint := &overlay.Outpoint{
				Txid:        *txid,
				OutputIndex: gaspTx.OutputIndex,
			}
			s.tempGraphNodeRefs[newGraphOutpoint.String()] = newGraphNode
		} else {
			return engine.ErrMissingInput
		}
		return nil
	}
}

func (s *overlayGASPStorage) ValidateGraphAnchor(ctx context.Context, graphID *overlay.Outpoint) error {
	if rootNode, ok := s.tempGraphNodeRefs[graphID.String()]; !ok {
		return engine.ErrMissingInput
	} else if beef, err := s.getBEEFForNode(ctx, rootNode); err != nil {
		return err
	} else if tx, err := transaction.NewTransactionFromBEEF(beef); err != nil {
		return err
	} else if valid, err := spv.Verify(tx, s.Engine.ChainTracker, nil); err != nil {
		return err
	} else if !valid {
		return errors.New("graph anchor is not a valid transaction")
	} else if beefs, err := s.computeOrderedBEEFsForGraph(ctx, graphID); err != nil {
		return err
	} else {
		coins := make(map[string]struct{})
		for _, beefBytes := range beefs {
			previousCoins := make([]uint32, 0)
			if tx, err := transaction.NewTransactionFromBEEF(beefBytes); err != nil {
				return err
			} else {
				for vin, input := range tx.Inputs {
					outpoint := &overlay.Outpoint{
						Txid:        *input.SourceTXID,
						OutputIndex: input.SourceTxOutIndex,
					}
					if _, ok := coins[outpoint.String()]; ok {
						previousCoins = append(previousCoins, uint32(vin))
					}
				}
				if admit, err := s.Engine.Managers[s.Topic].IdentifyAdmissableOutputs(ctx, beef, previousCoins); err != nil {
					return err
				} else {
					for _, vout := range admit.OutputsToAdmit {
						outpoint := &overlay.Outpoint{
							Txid:        *tx.TxID(),
							OutputIndex: vout,
						}
						coins[outpoint.String()] = struct{}{}
					}
				}
			}
		}
		if _, ok := coins[graphID.String()]; !ok {
			return errors.New("graph did not result in topical admittance of the root node. rejecting")
		}
		return nil
	}
}

func (s *overlayGASPStorage) DiscardGraph(ctx context.Context, graphID *overlay.Outpoint) {
	for nodeId, graphRef := range s.tempGraphNodeRefs {
		if graphRef.GraphID.String() == graphID.String() {
			delete(s.tempGraphNodeRefs, nodeId)
		}
	}
}

func (s *overlayGASPStorage) FinalizeGraph(ctx context.Context, graphID *overlay.Outpoint) error {
	if beefs, err := s.computeOrderedBEEFsForGraph(ctx, graphID); err != nil {
		return err
	} else {
		for _, beef := range beefs {
			if _, err := s.Engine.Submit(
				ctx,
				overlay.TaggedBEEF{
					Topics: []string{s.Topic},
					Beef:   beef,
				},
				engine.SubmitModeHistorical,
				nil,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *overlayGASPStorage) computeOrderedBEEFsForGraph(ctx context.Context, graphID *overlay.Outpoint) ([][]byte, error) {
	beefs := make([][]byte, 0)
	var hydrator func(node *GraphNode) error
	hydrator = func(node *GraphNode) error {
		if currentBeef, err := s.getBEEFForNode(ctx, node); err != nil {
			return err
		} else {
			if slices.IndexFunc(beefs, func(beef []byte) bool {
				return bytes.Equal(beef, currentBeef)
			}) == -1 {
				beefs = append([][]byte{currentBeef}, beefs...)
			}
			for _, child := range node.Children {
				if err := hydrator(child); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if foundRoot, ok := s.tempGraphNodeRefs[graphID.String()]; !ok {
		return nil, errors.New("unable to find root node in graph for finalization")
	} else if err := hydrator(foundRoot); err != nil {
		return nil, err
	} else {
		return beefs, nil
	}
}

func (s *overlayGASPStorage) getBEEFForNode(ctx context.Context, node *GraphNode) ([]byte, error) {
	var hydrator func(node *GraphNode) (*transaction.Transaction, error)
	hydrator = func(node *GraphNode) (*transaction.Transaction, error) {
		if tx, err := transaction.NewTransactionFromHex(node.RawTx); err != nil {
			return nil, err
		} else if len(node.Proof) > 0 {
			if tx.MerklePath, err = transaction.NewMerklePathFromHex(node.Proof); err != nil {
				return nil, err
			}
			return tx, nil
		} else {
			for vin, input := range tx.Inputs {
				outpoint := &overlay.Outpoint{
					Txid:        *input.SourceTXID,
					OutputIndex: input.SourceTxOutIndex,
				}
				if foundNode, ok := s.tempGraphNodeRefs[outpoint.String()]; !ok {
					return nil, errors.New("required input node for unproven parent not found in temporary graph store")
				} else if tx.Inputs[vin].SourceTransaction, err = hydrator(foundNode); err != nil {
					return nil, err
				}
			}
			return tx, nil
		}
	}
	if tx, err := hydrator(node); err != nil {
		return nil, err
	} else {
		return tx.AtomicBEEF(false)
	}
}
