package engine

import (
	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

// MerkleState represents the validation state of an output's merkle proof
type MerkleState uint8

const (
	MerkleStateUnmined MerkleState = iota
	MerkleStateValidated
	MerkleStateInvalidated
	MerkleStateImmutable
)

// String returns the string representation of the MerkleState
func (m MerkleState) String() string {
	switch m {
	case MerkleStateUnmined:
		return "Unmined"
	case MerkleStateValidated:
		return "Validated"
	case MerkleStateInvalidated:
		return "Invalidated"
	case MerkleStateImmutable:
		return "Immutable"
	default:
		return "Unknown"
	}
}

type Output struct {
	Outpoint        transaction.Outpoint
	Topic           string
	Script          *script.Script
	Satoshis        uint64
	Spent           bool
	OutputsConsumed []*transaction.Outpoint
	ConsumedBy      []*transaction.Outpoint
	BlockHeight     uint32
	BlockIdx        uint64
	Score           float64 // sort score for outputs. Usage is up to Storage implementation.
	Beef            []byte
	AncillaryTxids  []*chainhash.Hash
	AncillaryBeef   []byte
	MerkleRoot      *chainhash.Hash // Merkle root extracted from the merkle path
	MerkleState     MerkleState     // Validation state of the merkle proof
}
