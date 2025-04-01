package engine

import (
	"context"

	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/overlay/lookup"
	"github.com/bsv-blockchain/go-sdk/script"
)

type LookupService interface {
	OutputAdded(ctx context.Context, outpoint *overlay.Outpoint, outputScript *script.Script, topic string, blockHeight uint32, blockIndex uint64) error
	OutputSpent(ctx context.Context, outpoint *overlay.Outpoint, topic string) error
	OutputDeleted(ctx context.Context, outpoint *overlay.Outpoint, topic string) error
	OutputBlockHeightUpdated(ctx context.Context, outpoint *overlay.Outpoint, blockHeight uint32, blockIndex uint64) error
	Lookup(ctx context.Context, question *lookup.LookupQuestion) (*lookup.LookupAnswer, error)
	GetDocumentation() string
	GetMetaData() *overlay.MetaData
}
