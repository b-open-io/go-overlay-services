package gasp

import (
	"context"

	"github.com/bsv-blockchain/go-sdk/transaction"
)

type Storage interface {
	FindKnownUTXOs(ctx context.Context, since float64, limit uint32) ([]*Output, error)
	HasOutputs(ctx context.Context, outpoints []*transaction.Outpoint, topic string) (map[transaction.Outpoint]bool, error)
	HydrateGASPNode(ctx context.Context, graphID *transaction.Outpoint, outpoint *transaction.Outpoint, metadata bool) (*Node, error)
	FindNeededInputs(ctx context.Context, tx *Node) (*NodeResponse, error)
	AppendToGraph(ctx context.Context, tx *Node, spentBy *transaction.Outpoint) error
	ValidateGraphAnchor(ctx context.Context, graphID *transaction.Outpoint) error
	DiscardGraph(ctx context.Context, graphID *transaction.Outpoint) error
	FinalizeGraph(ctx context.Context, graphID *transaction.Outpoint) error
}
