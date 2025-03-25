package engine

import (
	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

type TopicManager interface {
	IdentifyAdmissableOutputs(beef *transaction.Beef, txid *chainhash.Hash, previousCoins []uint32) (overlay.AdmittanceInstructions, error)
	IdentifyNeededInputs(beef *transaction.Beef, txid *chainhash.Hash) ([]*overlay.Outpoint, error)
	GetDocumentation() string
	GetMetaData() *overlay.MetaData
}
