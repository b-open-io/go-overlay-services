package gasp

import (
	"context"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

type Storage interface {
	FindKnownUTXOs(ctx context.Context, since float64, limit uint32) ([]*Output, error)
	
	// HasOutputs returns the validation state of outputs for the given outpoints.
	// Returns a slice of *bool with the same length and order as the input outpoints:
	// - nil: output is unknown (not in storage)
	// - &true: output exists and has a valid merkle proof
	// - &false: output exists but has an invalid or missing merkle proof
	HasOutputs(ctx context.Context, outpoints []*transaction.Outpoint, topic string) ([]*bool, error)
	
	UpdateProof(ctx context.Context, txid *chainhash.Hash, proof *transaction.MerklePath) error
	HydrateGASPNode(ctx context.Context, graphID *transaction.Outpoint, outpoint *transaction.Outpoint, metadata bool) (*Node, error)
	FindNeededInputs(ctx context.Context, tx *Node) (*NodeResponse, error)
	AppendToGraph(ctx context.Context, tx *Node, spentBy *transaction.Outpoint) error
	ValidateGraphAnchor(ctx context.Context, graphID *transaction.Outpoint) error
	DiscardGraph(ctx context.Context, graphID *transaction.Outpoint) error
	FinalizeGraph(ctx context.Context, graphID *transaction.Outpoint) error
}
