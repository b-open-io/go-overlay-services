package commands_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4chain-ag/go-overlay-services/pkg/server/app/commands"
	"github.com/stretchr/testify/require"
)

type EngineProvider interface {
	StartGASPSync() error
}

// AlwaysSucceedsSync simulates a successful sync.
type AlwaysSucceedsSync struct{}

func (AlwaysSucceedsSync) StartGASPSync() error {
	return nil
}

// AlwaysFailsSync simulates a sync failure.
type AlwaysFailsSync struct{}

func (AlwaysFailsSync) StartGASPSync() error {
	return fmt.Errorf("simulated sync failure")
}

func TestStartGASPSyncHandler_Success(t *testing.T) {
	// Given:
	handler := commands.NewStartGASPSyncHandler(&AlwaysSucceedsSync{})
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// When:
	resp, err := ts.Client().Post(ts.URL, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Then:
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStartGASPSyncHandler_Failure(t *testing.T) {
	// Given:
	handler := commands.NewStartGASPSyncHandler(&AlwaysFailsSync{})
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// When:
	resp, err := ts.Client().Post(ts.URL, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Then:
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "FAILED")
}
