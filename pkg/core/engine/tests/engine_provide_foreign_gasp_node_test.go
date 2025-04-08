package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4chain-ag/go-overlay-services/pkg/core/engine"
)

// fakeStorage provides minimal stub for Storage interface
type fakeStorage struct {
	findOutputFunc func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error)
}

var errFakeStorage = errors.New("fakeStorage: method not implemented")

func (f fakeStorage) InsertOutput(ctx context.Context, utxo *engine.Output) error {
	return errFakeStorage
}

func (f fakeStorage) FindOutput(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
	return f.findOutputFunc(ctx, outpoint, topic, spent, includeBEEF)
}


func (f fakeStorage) FindOutputs(ctx context.Context, outpoints []*overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
	return nil, errFakeStorage
}

func (f fakeStorage) FindOutputsForTransaction(ctx context.Context, txid *chainhash.Hash, includeBEEF bool) ([]*engine.Output, error) {
	return nil, errFakeStorage
}

func (f fakeStorage) FindUTXOsForTopic(ctx context.Context, topic string, since uint32, includeBEEF bool) ([]*engine.Output, error) {
	return nil, errFakeStorage
}

func (f fakeStorage) DeleteOutput(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
	return errFakeStorage
}

func (f fakeStorage) DeleteOutputs(ctx context.Context, outpoints []*overlay.Outpoint, topic string) error {
	return errFakeStorage
}

func (f fakeStorage) MarkUTXOAsSpent(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
	return errFakeStorage
}

func (f fakeStorage) MarkUTXOsAsSpent(ctx context.Context, outpoints []*overlay.Outpoint, topic string) error {
	return errFakeStorage
}

func (f fakeStorage) UpdateConsumedBy(ctx context.Context, outpoint *overlay.Outpoint, topic string, consumedBy []*overlay.Outpoint) error {
	return errFakeStorage
}

func (f fakeStorage) UpdateTransactionBEEF(ctx context.Context, txid *chainhash.Hash, beef []byte) error {
	return errFakeStorage
}

func (f fakeStorage) UpdateOutputBlockHeight(ctx context.Context, outpoint *overlay.Outpoint, topic string, blockHeight uint32, blockIndex uint64, ancillaryBeef []byte) error {
	return errFakeStorage
}

func (f fakeStorage) InsertAppliedTransaction(ctx context.Context, tx *overlay.AppliedTransaction) error {
	return errFakeStorage
}

func (f fakeStorage) DoesAppliedTransactionExist(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
	return false, errFakeStorage
}

func createDummyBeef(t *testing.T) []byte {
	t.Helper()

	dummyLockingScript := script.Script{script.OpRETURN} 

	dummyTx := transaction.Transaction{
		Inputs: []*transaction.TransactionInput{},
		Outputs: []*transaction.TransactionOutput{
			{
				Satoshis:      1000,
				LockingScript: &dummyLockingScript,
			},
		},
	}

	beef, err := transaction.NewBeefFromTransaction(&dummyTx)
	require.NoError(t, err)

	serializedBytes, err := beef.AtomicBytes(dummyTx.TxID())
	require.NoError(t, err)

	return serializedBytes
}

func TestEngine_ProvideForeignGASPNode_Success(t *testing.T) {
	// given:
	ctx := context.Background()
	graphID := &overlay.Outpoint{}
	outpoint := &overlay.Outpoint{OutputIndex: 1}

	beefBytes := createDummyBeef(t)

	engine := &engine.Engine{
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Beef: beefBytes,
				}, nil
			},
		},
	}

	// when:
	node, err := engine.ProvideForeignGASPNode(ctx, graphID, outpoint)

	// then:
	require.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, graphID, node.GraphID)
	assert.Equal(t, outpoint.OutputIndex, node.OutputIndex)
}

func TestEngine_ProvideForeignGASPNode_MissingBeef_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	graphID := &overlay.Outpoint{}
	outpoint := &overlay.Outpoint{}
	engine := &engine.Engine{
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{}, nil // Missing Beef
			},
		},
	}

	// when:
	node, err := engine.ProvideForeignGASPNode(ctx, graphID, outpoint)

	// then:
	require.Error(t, err)
	assert.Nil(t, node)
}

func TestEngine_ProvideForeignGASPNode_CannotFindOutput_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	graphID := &overlay.Outpoint{}
	outpoint := &overlay.Outpoint{}
	engine := &engine.Engine{
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return nil, errors.New("forced error")
			},
		},
	}

	// when:
	node, err := engine.ProvideForeignGASPNode(ctx, graphID, outpoint)

	// then:
	require.Error(t, err)
	assert.Nil(t, node)
}

func TestEngine_ProvideForeignGASPNode_TransactionNotFound_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	graphID := &overlay.Outpoint{}
	outpoint := &overlay.Outpoint{}
	engine := &engine.Engine{
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Beef: []byte{0x00},
				}, nil
			},
		},
	}

	// when:
	node, err := engine.ProvideForeignGASPNode(ctx, graphID, outpoint)

	// then:
	require.Error(t, err)
	assert.Nil(t, node)
}
