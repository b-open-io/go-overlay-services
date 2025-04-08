package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/stretchr/testify/require"

	"github.com/4chain-ag/go-overlay-services/pkg/core/gasp/core"
)

type fakeGASPStorage struct {
	findKnownUTXOsFunc func(ctx context.Context, since uint32) ([]*overlay.Outpoint, error)
}

func (f fakeGASPStorage) FindKnownUTXOs(ctx context.Context, since uint32) ([]*overlay.Outpoint, error) {
	return f.findKnownUTXOsFunc(ctx, since)
}

func (f fakeGASPStorage) HydrateGASPNode(ctx context.Context, graphID *overlay.Outpoint, outpoint *overlay.Outpoint, metadata bool) (*core.GASPNode, error) {
	return nil, errors.New("not implemented")
}

func (f fakeGASPStorage) FindNeededInputs(ctx context.Context, tx *core.GASPNode) (*core.GASPNodeResponse, error) {
	return nil, errors.New("not implemented")
}

func (f fakeGASPStorage) AppendToGraph(ctx context.Context, tx *core.GASPNode, spentBy *chainhash.Hash) error {
	return errors.New("not implemented")
}

func (f fakeGASPStorage) ValidateGraphAnchor(ctx context.Context, graphID *overlay.Outpoint) error {
	return errors.New("not implemented")
}

func (f fakeGASPStorage) DiscardGraph(ctx context.Context, graphID *overlay.Outpoint) error {
	return errors.New("not implemented")
}

func (f fakeGASPStorage) FinalizeGraph(ctx context.Context, graphID *overlay.Outpoint) error {
	return errors.New("not implemented")
}

func TestGASP_GetInitialResponse_Success(t *testing.T) {
	// given:
	ctx := context.Background()
	request := &core.GASPInitialRequest{
		Version: 1,
		Since:   10,
	}
	expectedUTXOs := []*overlay.Outpoint{
		{OutputIndex: 1},
		{OutputIndex: 2},
	}
	gasp := core.NewGASP(core.GASPParams{
		Version:  ptr(1),
		Storage:  fakeGASPStorage{
			findKnownUTXOsFunc: func(ctx context.Context, since uint32) ([]*overlay.Outpoint, error) {
				return expectedUTXOs, nil
			},
		},
	})

	// when:
	resp, err := gasp.GetInitialResponse(ctx, request)

	// then:
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, expectedUTXOs, resp.UTXOList)
}

func TestGASP_GetInitialResponse_VersionMismatch_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	request := &core.GASPInitialRequest{
		Version: 99, // wrong version
		Since:   0,
	}
	gasp := core.NewGASP(core.GASPParams{
		Version: ptr(1),
		Storage: fakeGASPStorage{},
	})

	// when:
	resp, err := gasp.GetInitialResponse(ctx, request)

	// then:
	require.Error(t, err)
	require.Nil(t, resp)
	require.IsType(t, &core.GASPVersionMismatchError{}, err)
}

func TestGASP_GetInitialResponse_StorageFailure_ShouldReturnError(t *testing.T) {
	// given:
	ctx := context.Background()
	request := &core.GASPInitialRequest{
		Version: 1,
		Since:   0,
	}
	gasp := core.NewGASP(core.GASPParams{
		Version: ptr(1),
		Storage: fakeGASPStorage{
			findKnownUTXOsFunc: func(ctx context.Context, since uint32) ([]*overlay.Outpoint, error) {
				return nil, errors.New("forced storage error")
			},
		},
	})

	// when:
	resp, err := gasp.GetInitialResponse(ctx, request)

	// then:
	require.Error(t, err)
	require.Nil(t, resp)
}

func ptr(i int) *int {
	return &i
}
