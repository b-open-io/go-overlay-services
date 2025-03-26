package gasp

import (
	"github.com/4chain-ag/go-overlay-services/pkg/core/gasp/core"
	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
)

type GraphNode struct {
	Txid           *chainhash.Hash            `json:"txid"`
	GraphID        *overlay.Outpoint          `json:"graphID"`
	RawTx          string                     `json:"rawTx"`
	OutputIndex    uint32                     `json:"outputIndex"`
	SpentBy        *chainhash.Hash            `json:"spentBy"`
	Proof          string                     `json:"proof"`
	TxMetadata     string                     `json:"txMetadata"`
	OutputMetadata string                     `json:"outputMetadata"`
	Inputs         map[string]*core.GASPInput `json:"inputs"`
	Children       []*GraphNode               `json:"children"`
	Parent         *GraphNode                 `json:"parent"`
}
