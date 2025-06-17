package core

import (
	"fmt"

	"github.com/bsv-blockchain/go-sdk/transaction"
)

type GASPInitialRequest struct {
	Version int    `json:"version"`
	Since   uint32 `json:"since"`
}

type GASPInitialResponse struct {
	UTXOList []*transaction.Outpoint `json:"utxo_list"`
	Since    uint32                  `json:"since"`
}

// // MarshalJSON implements custom JSON marshaling for GASPInitialResponse
// func (g GASPInitialResponse) MarshalJSON() ([]byte, error) {
// 	type OutpointObj struct {
// 		Txid  string `json:"txid"`
// 		Index uint32 `json:"outputIndex"`
// 	}

// 	utxoList := make([]OutpointObj, len(g.UTXOList))
// 	for i, outpoint := range g.UTXOList {
// 		if outpoint != nil {
// 			utxoList[i] = OutpointObj{
// 				Txid:  outpoint.Txid.String(),
// 				Index: outpoint.Index,
// 			}
// 		}
// 	}

// 	return json.Marshal(&struct {
// 		UTXOList []OutpointObj `json:"utxo_list"`
// 		Since    uint32        `json:"since"`
// 	}{
// 		UTXOList: utxoList,
// 		Since:    g.Since,
// 	})
// }

// // UnmarshalJSON implements custom JSON unmarshalling for GASPInitialResponse
// func (g *GASPInitialResponse) UnmarshalJSON(data []byte) error {
// 	type OutpointObj struct {
// 		Txid  string `json:"txid"`
// 		Index uint32 `json:"index"`
// 	}

// 	aux := &struct {
// 		UTXOList []OutpointObj `json:"utxo_list"`
// 		Since    uint32        `json:"since"`
// 	}{}

// 	if err := json.Unmarshal(data, aux); err != nil {
// 		return err
// 	}

// 	g.Since = aux.Since
// 	g.UTXOList = make([]*transaction.Outpoint, len(aux.UTXOList))

// 	for i, obj := range aux.UTXOList {
// 		outpoint, err := transaction.OutpointFromString(fmt.Sprintf("%s.%d", obj.Txid, obj.Index))
// 		if err != nil {
// 			return err
// 		}
// 		g.UTXOList[i] = outpoint
// 	}

// 	return nil
// }

type GASPInitialReply struct {
	UTXOList []*transaction.Outpoint `json:"utxo_list"`
}

// // MarshalJSON implements custom JSON marshaling for GASPInitialReply
// func (g GASPInitialReply) MarshalJSON() ([]byte, error) {
// 	type OutpointObj struct {
// 		Txid  string `json:"txid"`
// 		Index uint32 `json:"index"`
// 	}

// 	utxoList := make([]OutpointObj, len(g.UTXOList))
// 	for i, outpoint := range g.UTXOList {
// 		if outpoint != nil {
// 			utxoList[i] = OutpointObj{
// 				Txid:  outpoint.Txid.String(),
// 				Index: outpoint.Index,
// 			}
// 		}
// 	}

// 	return json.Marshal(&struct {
// 		UTXOList []OutpointObj `json:"utxo_list"`
// 	}{
// 		UTXOList: utxoList,
// 	})
// }

// // UnmarshalJSON implements custom JSON unmarshalling for GASPInitialReply
// func (g *GASPInitialReply) UnmarshalJSON(data []byte) error {
// 	type OutpointObj struct {
// 		Txid  string `json:"txid"`
// 		Index uint32 `json:"index"`
// 	}

// 	aux := &struct {
// 		UTXOList []OutpointObj `json:"utxo_list"`
// 	}{}

// 	if err := json.Unmarshal(data, aux); err != nil {
// 		return err
// 	}

// 	g.UTXOList = make([]*transaction.Outpoint, len(aux.UTXOList))

// 	for i, obj := range aux.UTXOList {
// 		outpoint, err := transaction.OutpointFromString(fmt.Sprintf("%s.%d", obj.Txid, obj.Index))
// 		if err != nil {
// 			return err
// 		}
// 		g.UTXOList[i] = outpoint
// 	}

// 	return nil
// }

type GASPInput struct {
	Hash string `json:"hash"`
}

type GASPNode struct {
	GraphID        *transaction.Outpoint `json:"graphID"`
	RawTx          string                `json:"rawTx"`
	OutputIndex    uint32                `json:"outputIndex"`
	Proof          *string               `json:"proof"`
	TxMetadata     string                `json:"txMetadata"`
	OutputMetadata string                `json:"outputMetadata"`
	Inputs         map[string]*GASPInput `json:"inputs"`
	AncillaryBeef  []byte                `json:"ancillaryBeef"`
}

type GASPNodeResponseData struct {
	Metadata bool `json:"metadata"`
}

type GASPNodeResponse struct {
	RequestedInputs map[string]*GASPNodeResponseData `json:"requestedInputs"`
}

type GASPVersionMismatchError struct {
	Message        string `json:"message"`
	Code           string `json:"code"`
	CurrentVersion int    `json:"currentVersion"`
	ForeignVersion int    `json:"foreignVersion"`
}

func (e *GASPVersionMismatchError) Error() string {
	return e.Message
}

func NewGASPVersionMismatchError(currentVersion int, foreignVersion int) *GASPVersionMismatchError {
	return &GASPVersionMismatchError{
		Message:        fmt.Sprintf("GASP version mismatch. Current version: %d, foreign version: %d", currentVersion, foreignVersion),
		Code:           "ERR_GASP_VERSION_MISMATCH",
		CurrentVersion: currentVersion,
		ForeignVersion: foreignVersion,
	}
}
