package engine

import (
	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

type TopicContext struct {
	Tx           *transaction.Transaction
	Outputs      map[uint32]*Output
	Inputs       map[uint32]*Output
	Dependencies map[string]struct{}
	Admit        overlay.AdmittanceInstructions
}

// type DependencyLoader func(outpoint *overlay.Outpoint) (*Output, error)
type TopicManager interface {
	IdentifyAdmissableOutputs(beef *transaction.Beef, txid *chainhash.Hash, previousCoins map[uint32]*Output) (overlay.AdmittanceInstructions, error)
	IdentifyNeededInputs(beef *transaction.Beef) ([]*overlay.Outpoint, error)
	GetDependencies() []string
	GetDocumentation() string
	GetMetaData() *overlay.MetaData
}
