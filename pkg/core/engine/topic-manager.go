package engine

import (
	"context"

	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

type TopicManager interface {
	IdentifyAdmissibleOutputs(ctx context.Context, beef []byte, previousCoins []uint32) (overlay.AdmittanceInstructions, error)
	IdentifyNeededInputs(ctx context.Context, beef []byte) ([]*transaction.Outpoint, error)
	GetDocumentation() string
	GetMetaData() *overlay.MetaData
}
