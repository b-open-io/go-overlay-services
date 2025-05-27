package gasp_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4chain-ag/go-overlay-services/pkg/core/gasp/core"
	"github.com/bsv-blockchain/go-sdk/gasp"
	"github.com/stretchr/testify/require"
)

func TestGASPRemote_GetInitialResponse(t *testing.T) {
	t.Run("should send a request and return a valid response", func(t *testing.T) {
		// given
		mockRequest := &gasp.GASPInitialRequest{Version: 1, Since: 0}
		mockResponse := &gasp.GASPInitialResponse{
			UTXOList: []gasp.OutPoint{{TxID: "txid1", Vout: 0}},
			Since:    1234567890,
		}
		
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/requestSyncResponse", r.URL.Path)
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			require.Equal(t, "tm_test", r.Header.Get("X-BSV-Topic"))
			
			// Verify request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var receivedRequest gasp.GASPInitialRequest
			require.NoError(t, json.Unmarshal(body, &receivedRequest))
			require.Equal(t, mockRequest.Version, receivedRequest.Version)
			require.Equal(t, mockRequest.Since, receivedRequest.Since)
			
			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()
		
		sut := core.NewGASPRemote(server.URL, "tm_test")
		
		// when
		response, err := sut.GetInitialResponse(context.Background(), mockRequest)
		
		// then
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, mockResponse.Since, response.Since)
		require.Len(t, response.UTXOList, 1)
		require.Equal(t, "txid1", response.UTXOList[0].TxID)
		require.Equal(t, uint32(0), response.UTXOList[0].Vout)
	})
	
	t.Run("should throw an error if the response status is not OK", func(t *testing.T) {
		// given
		mockRequest := &gasp.GASPInitialRequest{Version: 1, Since: 0}
		
		// Create test server that returns 500
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		
		sut := core.NewGASPRemote(server.URL, "tm_test")
		
		// when
		response, err := sut.GetInitialResponse(context.Background(), mockRequest)
		
		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "HTTP error! Status: 500")
		require.Nil(t, response)
	})
	
	t.Run("should throw an error if the response format is invalid", func(t *testing.T) {
		// given
		mockRequest := &gasp.GASPInitialRequest{Version: 1, Since: 0}
		invalidResponse := map[string]string{"invalid": "data"}
		
		// Create test server that returns invalid response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(invalidResponse)
		}))
		defer server.Close()
		
		sut := core.NewGASPRemote(server.URL, "tm_test")
		
		// when
		response, err := sut.GetInitialResponse(context.Background(), mockRequest)
		
		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid response format")
		require.Nil(t, response)
	})
}

func TestGASPRemote_RequestNode(t *testing.T) {
	t.Run("should send a request and return a valid GASPNode", func(t *testing.T) {
		// given
		graphID := "graphID1"
		txid := "txid1"
		outputIndex := uint32(0)
		metadata := true
		mockResponse := &gasp.GASPNode{
			GraphID:        graphID,
			RawTX:          []byte("rawTxData"),
			OutputIndex:    outputIndex,
			Proof:          []byte("proofData"),
			TXMetadata:     "txMetadata",
			OutputMetadata: "outputMetadata",
			Inputs:         map[string]*gasp.GASPNode{},
		}
		
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/requestForeignGASPNode", r.URL.Path)
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			// Verify request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var receivedRequest map[string]interface{}
			require.NoError(t, json.Unmarshal(body, &receivedRequest))
			require.Equal(t, graphID, receivedRequest["graphID"])
			require.Equal(t, txid, receivedRequest["txid"])
			require.Equal(t, float64(outputIndex), receivedRequest["outputIndex"])
			require.Equal(t, metadata, receivedRequest["metadata"])
			
			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()
		
		sut := core.NewGASPRemote(server.URL, "tm_test")
		
		// when
		response, err := sut.RequestNode(context.Background(), graphID, txid, outputIndex, metadata)
		
		// then
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, graphID, response.GraphID)
		require.Equal(t, []byte("rawTxData"), response.RawTX)
		require.Equal(t, outputIndex, response.OutputIndex)
		require.Equal(t, []byte("proofData"), response.Proof)
		require.Equal(t, "txMetadata", response.TXMetadata)
		require.Equal(t, "outputMetadata", response.OutputMetadata)
	})
	
	t.Run("should throw an error if the response status is not OK", func(t *testing.T) {
		// given
		graphID := "graphID1"
		txid := "txid1"
		outputIndex := uint32(0)
		metadata := true
		
		// Create test server that returns 500
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		
		sut := core.NewGASPRemote(server.URL, "tm_test")
		
		// when
		response, err := sut.RequestNode(context.Background(), graphID, txid, outputIndex, metadata)
		
		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "HTTP error! Status: 500")
		require.Nil(t, response)
	})
	
	t.Run("should throw an error if the response format is invalid", func(t *testing.T) {
		// given
		graphID := "graphID1"
		txid := "txid1"
		outputIndex := uint32(0)
		metadata := true
		invalidResponse := map[string]string{"invalid": "data"}
		
		// Create test server that returns invalid response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(invalidResponse)
		}))
		defer server.Close()
		
		sut := core.NewGASPRemote(server.URL, "tm_test")
		
		// when
		response, err := sut.RequestNode(context.Background(), graphID, txid, outputIndex, metadata)
		
		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid response format")
		require.Nil(t, response)
	})
	
	t.Run("should handle HTTP errors properly", func(t *testing.T) {
		// given
		graphID := "graphID1"
		txid := "txid1"
		outputIndex := uint32(0)
		metadata := true
		
		// Create test server that returns various HTTP errors
		testCases := []struct {
			name       string
			statusCode int
		}{
			{"Bad Request", http.StatusBadRequest},
			{"Unauthorized", http.StatusUnauthorized},
			{"Not Found", http.StatusNotFound},
			{"Service Unavailable", http.StatusServiceUnavailable},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.statusCode)
				}))
				defer server.Close()
				
				sut := core.NewGASPRemote(server.URL, "tm_test")
				
				// when
				response, err := sut.RequestNode(context.Background(), graphID, txid, outputIndex, metadata)
				
				// then
				require.Error(t, err)
				require.Contains(t, err.Error(), "HTTP error!")
				require.Nil(t, response)
			})
		}
	})
}

func TestGASPRemote_RequestValidation(t *testing.T) {
	t.Run("should validate initial response fields", func(t *testing.T) {
		// given
		mockRequest := &gasp.GASPInitialRequest{Version: 1, Since: 0}
		
		// Response missing required fields
		invalidResponses := []struct {
			name     string
			response interface{}
		}{
			{
				name:     "missing UTXOList",
				response: map[string]interface{}{"Since": 123},
			},
			{
				name:     "missing Since",
				response: map[string]interface{}{"UTXOList": []interface{}{}},
			},
			{
				name:     "null response",
				response: nil,
			},
		}
		
		for _, tc := range invalidResponses {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tc.response)
				}))
				defer server.Close()
				
				sut := core.NewGASPRemote(server.URL, "tm_test")
				
				// when
				response, err := sut.GetInitialResponse(context.Background(), mockRequest)
				
				// then
				require.Error(t, err)
				require.Nil(t, response)
			})
		}
	})
	
	t.Run("should validate node response fields", func(t *testing.T) {
		// given
		graphID := "graphID1"
		txid := "txid1"
		outputIndex := uint32(0)
		metadata := true
		
		// Response missing required fields
		invalidResponses := []struct {
			name     string
			response interface{}
		}{
			{
				name:     "missing GraphID",
				response: map[string]interface{}{"RawTX": "data"},
			},
			{
				name:     "missing RawTX",
				response: map[string]interface{}{"GraphID": "id"},
			},
			{
				name:     "null response",
				response: nil,
			},
		}
		
		for _, tc := range invalidResponses {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tc.response)
				}))
				defer server.Close()
				
				sut := core.NewGASPRemote(server.URL, "tm_test")
				
				// when
				response, err := sut.RequestNode(context.Background(), graphID, txid, outputIndex, metadata)
				
				// then
				require.Error(t, err)
				require.Nil(t, response)
			})
		}
	})
}

// Mock implementation helper
type mockGASPRemote struct {
	getInitialResponseFunc func(ctx context.Context, request *gasp.GASPInitialRequest) (*gasp.GASPInitialResponse, error)
	requestNodeFunc        func(ctx context.Context, graphID, txid string, outputIndex uint32, metadata bool) (*gasp.GASPNode, error)
}

func (m *mockGASPRemote) GetInitialResponse(ctx context.Context, request *gasp.GASPInitialRequest) (*gasp.GASPInitialResponse, error) {
	if m.getInitialResponseFunc != nil {
		return m.getInitialResponseFunc(ctx, request)
	}
	return nil, errors.New("not implemented")
}

func (m *mockGASPRemote) RequestNode(ctx context.Context, graphID, txid string, outputIndex uint32, metadata bool) (*gasp.GASPNode, error) {
	if m.requestNodeFunc != nil {
		return m.requestNodeFunc(ctx, graphID, txid, outputIndex, metadata)
	}
	return nil, errors.New("not implemented")
}