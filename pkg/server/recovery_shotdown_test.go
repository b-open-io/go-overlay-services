package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/4chain-ag/go-overlay-services/pkg/server"
	"github.com/4chain-ag/go-overlay-services/pkg/server/config"
	"github.com/stretchr/testify/require"
)

// RecoveryMiddleware is a middleware that recovers from panics and returns a 500 status code.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Test_RecoveryMiddleware_ShouldHandlePanic(t *testing.T) {
	// Given
	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected error")
	})

	handler := RecoveryMiddleware(mux)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// When
	resp, err := http.Get(ts.URL + "/panic")

	// Then
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func Test_GracefulShutdown_ShouldShutdownWithoutError(t *testing.T) {
	// Given
	cfg := config.DefaultConfig()
	opts := []server.HTTPOption{
		server.WithConfig(cfg),
	}
	httpAPI := server.New(opts...)

	// When
	done := httpAPI.StartWithGracefulShutdown()
	time.Sleep(200 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)

	// Then
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shutdown gracefully in time")
	}
}
