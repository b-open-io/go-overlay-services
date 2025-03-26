package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/idempotency"

	config "github.com/4chain-ag/go-overlay-services/pkg/appconfig"
	"github.com/4chain-ag/go-overlay-services/pkg/server/app"
	"github.com/4chain-ag/go-overlay-services/pkg/server/app/jsonutil"
	"github.com/4chain-ag/go-overlay-services/pkg/server/mongo"
)

// HTTPOption defines a functional option for configuring an HTTP server.
// These options allow for flexible setup of middlewares and configurations.
type HTTPOption func(*HTTP) error

// WithMiddleware adds net/http-style middleware to the HTTP server.
func WithMiddleware(f func(http.Handler) http.Handler) HTTPOption {
	return func(h *HTTP) error {
		h.middlewares = append(h.middlewares, adaptor.HTTPMiddleware(f))
		return nil
	}
}

<<<<<<< HEAD
// WithConfig sets the configuration for the HTTP server.
=======
// WithFiberMiddleware adds a Fiber-style middleware to the HTTP server.
func WithFiberMiddleware(m fiber.Handler) HTTPOption {
	return func(h *HTTP) {
		h.middlewares = append(h.middlewares, m)
	}
}

// WithConfig sets the HTTP server configuration.
>>>>>>> 38cbff3 (adding CROS support)
func WithConfig(cfg *config.Config) HTTPOption {
	return func(h *HTTP) error {
		h.cfg = cfg
		return nil
	}
}

// WithMongo sets the MongoDB client for the HTTP server based on the configuration.
func WithMongo() HTTPOption {
	return func(h *HTTP) error {
		if h.cfg == nil || h.cfg.Mongo.URI == "" {
			return nil
		}
		client, err := mongo.Connect(h.cfg)
		if err != nil {
			return fmt.Errorf("MongoDB connect failed: %w", err)
		}
		h.mongo = client
		return nil
	}
}

// HTTP manages the Fiber server and its configuration.
type HTTP struct {
	middlewares []fiber.Handler
	app         *fiber.App
	cfg         *config.Config
	mongo       *mongo.Client
}

// New returns an instance of the HTTP server and applies all specified functional options before starting it.
func New(opts ...HTTPOption) (*HTTP, error) {
	overlayAPI, err := app.New(NewNoopEngineProvider())
	if err != nil {
		return nil, fmt.Errorf("failed to create overlay API: %w", err)
	}

	http := &HTTP{
		app: fiber.New(fiber.Config{
			CaseSensitive: true,
			StrictRouting: true,
			ServerHeader:  "Overlay API",
			AppName:       "Overlay API v0.0.0",
		}),
		middlewares: []fiber.Handler{
			idempotency.New(),
			cors.New(),
		},
	}

	for _, o := range opts {
		if err := o(http); err != nil {
			return nil, err
		}
	}

	for _, m := range http.middlewares {
		http.app.Use(m)
	}

	// Routes...
	api := http.app.Group("/api")
	v1 := api.Group("/v1")

	// Non-Admin:
	v1.Post("/submit", adaptor.HTTPHandlerFunc(overlayAPI.Commands.SubmitTransactionHandler.Handle))
	v1.Get("/topic-managers", adaptor.HTTPHandlerFunc(overlayAPI.Queries.TopicManagerDocumentationHandler.Handle))
	v1.Post("/request-foreign-gasp-node", adaptor.HTTPHandlerFunc(overlayAPI.Commands.RequestForeignGASPNodeHandler.Handle))

	// Admin:
	admin := v1.Group("/admin", adaptor.HTTPMiddleware(AdminAuth(http.cfg.AdminBearerToken)))
	admin.Post("/advertisements-sync", adaptor.HTTPHandlerFunc(overlayAPI.Commands.SyncAdvertismentsHandler.Handle))
	admin.Post("/start-gasp-sync", adaptor.HTTPHandlerFunc(overlayAPI.Commands.StartGASPSyncHandler.Handle))

	return http, nil
}

// SocketAddr builds the address string for binding.
func (h *HTTP) SocketAddr() string {
	return fmt.Sprintf("%s:%d", h.cfg.Addr, h.cfg.Port)
}

// ListenAndServe starts the Fiber app using the configured socket address.
func (h *HTTP) ListenAndServe() error {
	if err := h.app.Listen(h.SocketAddr()); err != nil {
		return fmt.Errorf("http server: fiber app listen failed: %w", err)
	}
	return nil
}

// AdminAuth is a middleware that checks the Authorization header for a valid Bearer token.
// protects the HTTP server from unauthorized access.
// It checks for a Bearer token in the Authorization header and compares it to the expected value.
func AdminAuth(expectedToken string) func(http.Handler) http.Handler {
	type AuthorizationFailureResponse struct {
		Status  string `json:"error"`
		Message string `json:"message"`
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				jsonutil.SendHTTPResponse(w, http.StatusUnauthorized, AuthorizationFailureResponse{
					Status:  http.StatusText(http.StatusUnauthorized),
					Message: "Missing Authorization header in the request",
				})
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				jsonutil.SendHTTPResponse(w, http.StatusUnauthorized, AuthorizationFailureResponse{
					Status:  http.StatusText(http.StatusUnauthorized),
					Message: "Missing Authorization header Bearer token value",
				})
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			if token != expectedToken {
				jsonutil.SendHTTPResponse(w, http.StatusForbidden, AuthorizationFailureResponse{
					Status:  http.StatusText(http.StatusForbidden),
					Message: "Invalid Bearer token value",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// App exposes the underlying Fiber app.
func (h *HTTP) App() *fiber.App {
	return h.app
}
