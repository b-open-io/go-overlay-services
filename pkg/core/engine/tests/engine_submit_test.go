package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/4chain-ag/go-overlay-services/pkg/core/engine"
	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/stretchr/testify/require"
)

func TestEngine_Submit_Success(t *testing.T) {
	// given:
	ctx := context.Background()

	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
					}, nil
				},
			},
		},
		Storage: fakeStorage{
			deleteOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
				return nil
			},
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{{}}, nil
			},
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return false, nil
			},
			markUTXOsAsSpentFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spendTxid *chainhash.Hash) error {
				return nil
			},
			insertOutputFunc: func(ctx context.Context, output *engine.Output) error {
				return nil
			},
			insertAppliedTransactionFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) error {
				return nil
			},
		},
		ChainTracker: fakeChainTracker{
			isValidRootForHeight: func(root *chainhash.Hash, height uint32) (bool, error) {
				return true, nil
			},
		},
	}

	taggedBEEF := overlay.TaggedBEEF{
		Topics: []string{"test-topic"},
		Beef:   createDummyBEEF(t),
	}

	expectedSteak := overlay.Steak{
		"test-topic": &overlay.AdmittanceInstructions{
			OutputsToAdmit: []uint32{0},
			CoinsRemoved:   []uint32{0},
		},
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.NoError(t, err)
	require.Equal(t, expectedSteak, steak)
}

func TestEngine_Submit_InvalidBeef_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
					}, nil
				},
			},
		},
		Storage:      fakeStorage{},
		ChainTracker: fakeChainTracker{},
	}

	taggedBEEF := overlay.TaggedBEEF{
		Topics: []string{"test-topic"},
		Beef:   []byte{0xFF}, // invalid beef
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid-version") // temp fix for SPV failure Submit need to be fixed by wrapping the error to use ErrorIs
	require.Nil(t, steak)
}

func TestEngine_Submit_SPVFail_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
					}, nil
				},
			},
		},
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Outpoint: *outpoint,
					Satoshis: 1000,
					Script:   &script.Script{script.OpTRUE},
				}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{
					{
						Outpoint: *outpoints[0],
						Satoshis: 1000,
						Script:   &script.Script{script.OpTRUE},
					},
				}, nil
			},
		},
		ChainTracker: fakeChainTrackerSPVFail{},
	}

	taggedBEEF := overlay.TaggedBEEF{
		Topics: []string{"test-topic"},
		Beef:   createDummyBeefWithInputs(t),
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.Error(t, err)
	require.Equal(t, err.Error(), "input 0 has no source transaction") // temp fix for SPV failure Submit need to be fixed by wrapping the error to use ErrorIs
	require.Nil(t, steak)
}

func TestEngine_Submit_DuplicateTransaction_ShouldReturnEmptySteak(t *testing.T) {
	// given:
	ctx := context.Background()
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{},
		},
		Storage: fakeStorage{
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return true, nil
			},
		},
		ChainTracker: fakeChainTracker{
			isValidRootForHeight: func(root *chainhash.Hash, height uint32) (bool, error) {
				return true, nil
			},
		},
	}
	taggedBEEF := overlay.TaggedBEEF{
		Topics: []string{"test-topic"},
		Beef:   createDummyBEEF(t),
	}

	expectedSteak := overlay.Steak{
		"test-topic": &overlay.AdmittanceInstructions{
			OutputsToAdmit: nil,
		},
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.NoError(t, err)
	require.Equal(t, expectedSteak, steak)
}

func TestEngine_Submit_MissingTopic_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	sut := &engine.Engine{
		Managers:     map[string]engine.TopicManager{},
		Storage:      fakeStorage{},
		ChainTracker: fakeChainTracker{},
	}
	taggedBEEF := overlay.TaggedBEEF{
		Topics: []string{"unknown-topic"},
		Beef:   createDummyBEEF(t),
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.ErrorIs(t, err, engine.ErrUnknownTopic)
	require.Nil(t, steak)
}

func TestEngine_Submit_BroadcastFails_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
					}, nil
				},
			},
		},
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{{}}, nil
			},
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return false, nil
			},
			markUTXOsAsSpentFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spendTxid *chainhash.Hash) error {
				return nil
			},
		},
		ChainTracker: fakeChainTracker{
			verifyFunc: func(tx *transaction.Transaction, options ...any) (bool, error) {
				return true, nil
			},
			isValidRootForHeight: func(root *chainhash.Hash, height uint32) (bool, error) {
				return true, nil
			},
		},
		Broadcaster: fakeBroadcasterFail{
			broadcastFunc: func(tx *transaction.Transaction) (*transaction.BroadcastSuccess, *transaction.BroadcastFailure) {
				return nil, &transaction.BroadcastFailure{Description: "forced failure for testing"}
			},
			broadcastCtxFunc: func(ctx context.Context, tx *transaction.Transaction) (*transaction.BroadcastSuccess, *transaction.BroadcastFailure) {
				return nil, &transaction.BroadcastFailure{Description: "forced failure for testing"}
			},
		},
	}

	taggedBEEF := overlay.TaggedBEEF{
		Topics: []string{"test-topic"},
		Beef:   createDummyBEEF(t),
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.Error(t, err)
	require.Nil(t, steak)
	require.EqualError(t, err, "forced failure for testing")
}

func TestEngine_Submit_CoinsRetained(t *testing.T) {
	// Test when identifyAdmissibleOutputs returns coinsToRetain
	// Verify outputs are marked spent but not deleted
	ctx := context.Background()
	taggedBEEF, prevTxID := createDummyValidTaggedBEEF(t)
	
	markSpentCalled := false
	deleteOutputCalled := false
	outputSpentCalled := false
	outputNoLongerRetainedCalled := false
	
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
						CoinsToRetain:  []uint32{0}, // Retain the first input
					}, nil
				},
			},
		},
		LookupServices: map[string]engine.LookupService{
			"test-topic": fakeLookupService{
				outputSpentFunc: func(ctx context.Context, notification overlay.OutputSpentNotification) error {
					outputSpentCalled = true
					require.Equal(t, "none", notification.Mode)
					require.Equal(t, prevTxID.String(), notification.Txid)
					require.Equal(t, uint32(0), notification.OutputIndex)
					require.Equal(t, "test-topic", notification.Topic)
					return nil
				},
				outputNoLongerRetainedInHistoryFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
					outputNoLongerRetainedCalled = true
					return nil
				},
			},
		},
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Topic:    "test-topic",
				}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Topic:    "test-topic",
				}}, nil
			},
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return false, nil
			},
			markUTXOsAsSpentFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spendTxid *chainhash.Hash) error {
				markSpentCalled = true
				require.Len(t, outpoints, 1)
				require.Equal(t, prevTxID.String(), outpoints[0].Txid.String())
				require.Equal(t, uint32(0), outpoints[0].OutputIndex)
				return nil
			},
			deleteOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
				deleteOutputCalled = true
				return nil
			},
			insertOutputFunc: func(ctx context.Context, output *engine.Output) error {
				return nil
			},
			insertAppliedTransactionFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) error {
				return nil
			},
			updateConsumedByFunc: func(ctx context.Context, outpoint *overlay.Outpoint, spentBy *overlay.Outpoint, topic string) error {
				// Verify consumption tracking for retained coins
				require.Equal(t, prevTxID.String(), outpoint.Txid.String())
				require.Equal(t, uint32(0), outpoint.OutputIndex)
				require.NotNil(t, spentBy)
				return nil
			},
		},
		ChainTracker: fakeChainTracker{
			verifyFunc: func(tx *transaction.Transaction, options ...any) (bool, error) {
				return true, nil
			},
			isValidRootForHeight: func(root *chainhash.Hash, height uint32) (bool, error) {
				return true, nil
			},
		},
	}
	
	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)
	
	// then:
	require.NoError(t, err)
	require.NotNil(t, steak)
	require.True(t, markSpentCalled, "UTXO should be marked as spent")
	require.False(t, deleteOutputCalled, "UTXO should NOT be deleted when retained")
	require.True(t, outputSpentCalled, "Lookup service should be notified of spent output")
	require.False(t, outputNoLongerRetainedCalled, "outputNoLongerRetainedInHistory should NOT be called for retained coins")
	
	// Verify the steak contains the retained coin info
	require.Contains(t, steak, "test-topic")
	require.Contains(t, steak["test-topic"].CoinsToRetain, uint32(0))
}

func TestEngine_Submit_CoinsNotRetained(t *testing.T) {
	// Test when previous UTXOs were not retained by the topic manager
	// Verify deleteUTXODeep is called
	ctx := context.Background()
	taggedBEEF, prevTxID := createDummyValidTaggedBEEF(t)
	
	markSpentCalled := false
	deleteOutputCalled := false
	outputSpentCalled := false
	outputNoLongerRetainedCalled := false
	
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
						CoinsToRetain:  []uint32{}, // No coins retained
					}, nil
				},
			},
		},
		LookupServices: map[string]engine.LookupService{
			"test-topic": fakeLookupService{
				outputSpentFunc: func(ctx context.Context, notification overlay.OutputSpentNotification) error {
					outputSpentCalled = true
					return nil
				},
				outputNoLongerRetainedInHistoryFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
					outputNoLongerRetainedCalled = true
					require.Equal(t, prevTxID.String(), outpoint.Txid.String())
					require.Equal(t, uint32(0), outpoint.OutputIndex)
					require.Equal(t, "test-topic", topic)
					return nil
				},
			},
		},
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Topic:    "test-topic",
				}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Topic:    "test-topic",
				}}, nil
			},
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return false, nil
			},
			markUTXOsAsSpentFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spendTxid *chainhash.Hash) error {
				markSpentCalled = true
				return nil
			},
			deleteOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
				deleteOutputCalled = true
				require.Equal(t, prevTxID.String(), outpoint.Txid.String())
				require.Equal(t, uint32(0), outpoint.OutputIndex)
				require.Equal(t, "test-topic", topic)
				return nil
			},
			insertOutputFunc: func(ctx context.Context, output *engine.Output) error {
				return nil
			},
			insertAppliedTransactionFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) error {
				return nil
			},
		},
		ChainTracker: fakeChainTracker{
			verifyFunc: func(tx *transaction.Transaction, options ...any) (bool, error) {
				return true, nil
			},
			isValidRootForHeight: func(root *chainhash.Hash, height uint32) (bool, error) {
				return true, nil
			},
		},
	}
	
	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)
	
	// then:
	require.NoError(t, err)
	require.NotNil(t, steak)
	require.True(t, markSpentCalled, "UTXO should be marked as spent")
	require.True(t, deleteOutputCalled, "UTXO should be deleted when not retained")
	require.True(t, outputSpentCalled, "Lookup service should be notified of spent output")
	require.True(t, outputNoLongerRetainedCalled, "Lookup service should be notified that output is no longer retained")
	
	// Verify the steak shows coins were removed
	require.Contains(t, steak, "test-topic")
	require.Empty(t, steak["test-topic"].CoinsToRetain)
}

func TestEngine_Submit_OutputInsertFails_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	taggedBEEF, prevTxID := createDummyValidTaggedBEEF(t)
	expectedErr := errors.New("insert-failed")

	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
					}, nil
				},
			},
		},
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Script:   &script.Script{script.OpTRUE},
					Topic:    "test-topic",
				}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{
					{
						Outpoint: overlay.Outpoint{
							Txid:        *prevTxID,
							OutputIndex: 0,
						},
						Satoshis: 1000,
						Script:   &script.Script{script.OpTRUE},
						Topic:    "test-topic",
					},
				}, nil
			},
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return false, nil
			},
			markUTXOsAsSpentFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spendTxid *chainhash.Hash) error {
				return nil
			},
			insertOutputFunc: func(ctx context.Context, output *engine.Output) error {
				return expectedErr
			},
			deleteOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic string) error {
				return nil
			},
		},
		ChainTracker: fakeChainTracker{},
	}

	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)

	// then:
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, steak)
}

func TestEngine_Submit_AppliedTransactionInsertionVerification(t *testing.T) {
	// Test that applied transaction is inserted correctly
	ctx := context.Background()
	taggedBEEF, prevTxID := createDummyValidTaggedBEEF(t)
	
	appliedTxInserted := false
	var insertedAppliedTx *overlay.AppliedTransaction
	
	sut := &engine.Engine{
		Managers: map[string]engine.TopicManager{
			"test-topic": fakeManager{
				identifyAdmissableOutputsFunc: func(ctx context.Context, beef []byte, previousCoins map[uint32]*transaction.TransactionOutput) (overlay.AdmittanceInstructions, error) {
					return overlay.AdmittanceInstructions{
						OutputsToAdmit: []uint32{0},
					}, nil
				},
			},
		},
		Storage: fakeStorage{
			findOutputFunc: func(ctx context.Context, outpoint *overlay.Outpoint, topic *string, spent *bool, includeBEEF bool) (*engine.Output, error) {
				return &engine.Output{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Topic:    "test-topic",
				}, nil
			},
			findOutputsFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spent *bool, includeBEEF bool) ([]*engine.Output, error) {
				return []*engine.Output{{
					Outpoint: overlay.Outpoint{
						Txid:        *prevTxID,
						OutputIndex: 0,
					},
					Satoshis: 1000,
					Topic:    "test-topic",
				}}, nil
			},
			doesAppliedTransactionExistFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) (bool, error) {
				return false, nil
			},
			markUTXOsAsSpentFunc: func(ctx context.Context, outpoints []*overlay.Outpoint, topic string, spendTxid *chainhash.Hash) error {
				return nil
			},
			insertOutputFunc: func(ctx context.Context, output *engine.Output) error {
				return nil
			},
			insertAppliedTransactionFunc: func(ctx context.Context, tx *overlay.AppliedTransaction) error {
				appliedTxInserted = true
				insertedAppliedTx = tx
				// Verify the applied transaction details
				require.NotNil(t, tx)
				require.Equal(t, "test-topic", tx.Topic)
				require.NotNil(t, tx.Txid)
				return nil
			},
		},
		ChainTracker: fakeChainTracker{
			verifyFunc: func(tx *transaction.Transaction, options ...any) (bool, error) {
				return true, nil
			},
			isValidRootForHeight: func(root *chainhash.Hash, height uint32) (bool, error) {
				return true, nil
			},
		},
	}
	
	// when:
	steak, err := sut.Submit(ctx, taggedBEEF, engine.SubmitModeCurrent, nil)
	
	// then:
	require.NoError(t, err)
	require.NotNil(t, steak)
	require.True(t, appliedTxInserted, "Applied transaction should be inserted")
	require.NotNil(t, insertedAppliedTx, "Inserted applied transaction should not be nil")
	require.Equal(t, "test-topic", insertedAppliedTx.Topic)
}
