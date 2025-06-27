package gasp

import (
	"context"

	"github.com/bsv-blockchain/go-sdk/transaction"
)

type Remote interface {
	GetInitialResponse(ctx context.Context, request *InitialRequest) (*InitialResponse, error)
	GetInitialReply(ctx context.Context, response *InitialResponse) (*InitialReply, error)
	RequestNode(ctx context.Context, graphID *transaction.Outpoint, outpoint *transaction.Outpoint, metadata bool) (*Node, error)
	SubmitNode(ctx context.Context, node *Node) (*NodeResponse, error)
}
