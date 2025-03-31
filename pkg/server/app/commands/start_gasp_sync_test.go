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

type alwaysSucceedsSync struct{}

func (alwaysSucceedsSync) StartGASPSync() error {
	return nil
}

type alwaysFailsSync struct{}

func (alwaysFailsSync) StartGASPSync() error {
	return fmt.Errorf("simulated sync failure")
}

func TestStartGASPSyncHandler_Success(t *testing.T) {
	// Given:
	handler := commands.NewStartGASPSyncHandler(&alwaysSucceedsSync{})
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// When:
	resp, err := ts.Client().Post(ts.URL, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Then:
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStartGASPSyncHandler_Failure(t *testing.T) {
	// Given:
	handler := commands.NewStartGASPSyncHandler(&alwaysFailsSync{})
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
