package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/4chain-ag/go-overlay-services/pkg/server"
)

func Test_RecoveryMiddleware_ShouldHandlePanic(t *testing.T) {
	// Given
	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected error")
	})

	handler := server.RecoveryMiddleware(mux)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// When
	resp, err := http.Get(ts.URL + "/panic")

	// Then
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
