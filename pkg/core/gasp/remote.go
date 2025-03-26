package gasp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/4chain-ag/go-overlay-services/pkg/core/gasp/core"
	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/util"
)

type OverlayGASPRemote struct {
	EndpointUrl string
	Topic       string
	HttpClient  util.HTTPClient
}

func (r *OverlayGASPRemote) InitialResponse(request *core.GASPInitialRequest) (*core.GASPInitialResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return nil, err
	} else if req, err := http.NewRequest("POST", r.EndpointUrl+"/requestSyncResponse", io.NopCloser(&buf)); err != nil {
		return nil, err
	} else {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-BSV-Topic", r.Topic)
		if resp, err := r.HttpClient.Do(req); err != nil {
			return nil, err
		} else {
			if resp.StatusCode != http.StatusOK {
				return nil, &util.HTTPError{
					StatusCode: resp.StatusCode,
					Err:        err,
				}
			}
			result := &core.GASPInitialResponse{}
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
}

func (r *OverlayGASPRemote) RequestNode(ctx context.Context, graphID *overlay.Outpoint, outpoint *overlay.Outpoint, metadata bool) (*core.GASPNode, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"graphID":     graphID,
		"txid":        outpoint.Txid.String(),
		"outputIndex": outpoint.OutputIndex,
		"metadata":    metadata,
	}); err != nil {
		return nil, err
	} else if req, err := http.NewRequest("POST", r.EndpointUrl+"/requestForeignGASPNode", io.NopCloser(&buf)); err != nil {
		return nil, err
	} else {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-BSV-Topic", r.Topic)
		if resp, err := r.HttpClient.Do(req); err != nil {
			return nil, err
		} else {
			if resp.StatusCode != http.StatusOK {
				return nil, &util.HTTPError{
					StatusCode: resp.StatusCode,
					Err:        err,
				}
			}
			result := &core.GASPNode{}
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
}

func (r *OverlayGASPRemote) InitialReplay(response *core.GASPInitialResponse) (*core.GASPInitialReply, error) {
	return nil, errors.New("not-implemented")
}

func (r *OverlayGASPRemote) SubmitNode(node *core.GASPNode) (*core.GASPNodeResponse, error) {
	return nil, errors.New("not-implemented")
}
