package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4chain-ag/go-overlay-services/pkg/server"
	"github.com/4chain-ag/go-overlay-services/pkg/server/app/commands"
	"github.com/4chain-ag/go-overlay-services/pkg/server/app/queries"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/stretchr/testify/require"
)

func startTestServer(t *testing.T) *httptest.Server {
	app := fiber.New()
	noopProvider := server.NewNoopEngineProvider()

	app.Post("/submit", commands.NewSubmitTransactionCommandHandler(noopProvider).Handle)
	app.Post("/admin/advertisements-sync", commands.NewSyncAdvertisementsHandler(noopProvider).Handle)
	app.Get("/topic-managers", queries.NewTopicManagerDocumentationHandler(noopProvider).Handle)

	return httptest.NewServer(adaptor.FiberApp(app))
}

func newTestRestyClient(ts *httptest.Server) *resty.Client {
	return resty.New().SetBaseURL(ts.URL)
}

func TestSubmitTransactionHandler_ShouldReturnOK(t *testing.T) {
	// Given
	ts := startTestServer(t)
	defer ts.Close()

	client := newTestRestyClient(ts)

	// When
	resp, err := client.R().Post("/submit")

	// Then
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestSyncAdvertisementsHandler_ShouldReturnOK(t *testing.T) {
	// Given
	ts := startTestServer(t)
	defer ts.Close()

	client := newTestRestyClient(ts)

	// When
	resp, err := client.R().Post("/admin/advertisements-sync")

	// Then
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestTopicManagerDocumentationHandler_ShouldReturnOK(t *testing.T) {
	// Given
	ts := startTestServer(t)
	defer ts.Close()

	client := newTestRestyClient(ts)

	// When
	resp, err := client.R().Get("/topic-managers")

	// Then
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}
