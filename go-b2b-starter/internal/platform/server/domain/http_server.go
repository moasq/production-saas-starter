package domain

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	config "github.com/moasq/go-b2b-starter/internal/platform/server/config"
)

type readinessDatabase interface {
	Ping(context.Context) error
}

// RuntimeReadiness validates the database contract using the request deadline.
type RuntimeReadiness func(context.Context) error

type HTTPServer struct {
	database         readinessDatabase
	runtimeReadiness RuntimeReadiness
	config           *config.Config
	router           *gin.Engine
	logger           logger.Logger
	namedMiddlewares map[string]MiddlewareFunc
}

func NewHTTPServer(
	config *config.Config,
	router *gin.Engine,
	log logger.Logger,
	database *pgxpool.Pool,
	runtimeReadiness RuntimeReadiness,
) Server {
	server := &HTTPServer{
		database:         database,
		runtimeReadiness: runtimeReadiness,
		config:           config,
		router:           router,
		logger:           log,
		namedMiddlewares: make(map[string]MiddlewareFunc),
	}

	server.setupMiddleware()
	return server
}

// Start initializes and starts the HTTP server
func (s *HTTPServer) Start() error {
	srv := s.createHTTPServer()
	s.setupHealthCheck()
	s.setupRootEndpoint()

	go s.startServer(srv)
	return s.handleGracefulShutdown(srv)
}

// RegisterRoutes registers route handlers with version support
func (s *HTTPServer) RegisterRoutes(registrar RouteRegistrar, prefix string, version ...string) {
	v := ""
	if len(version) > 0 {
		v = version[0]
	}

	group := s.router.Group(prefix)
	if v != "" {
		group = group.Group("/" + v)
	}

	// Register routes immediately instead of storing for later
	registrar(group, s)
}

// RegisterNamedMiddleware registers a named middleware for later use
func (s *HTTPServer) RegisterNamedMiddleware(name string, middleware MiddlewareFunc) {
	s.namedMiddlewares[name] = middleware
	s.logger.Info("Named middleware registered: " + name)
}

func (s *HTTPServer) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:              s.config.ServerAddress,
		Handler:           s.router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    s.config.MaxRequestSize,
	}
}

func (s *HTTPServer) startServer(srv *http.Server) {
	s.logger.Info("Starting server on " + s.config.ServerAddress)
	var err error

	if s.config.EnableTLS {
		err = srv.ListenAndServeTLS(
			s.config.TLSCertPath,
			s.config.TLSKeyPath,
		)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("Failed to start server", logger.Fields{"error": err})
	}
}

func (s *HTTPServer) handleGracefulShutdown(srv *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	<-quit
	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down server: %w", err)
	}

	s.logger.Info("Server exited gracefully")
	return nil
}

// Get implements the MiddlewareResolver interface
func (s *HTTPServer) Get(name string) gin.HandlerFunc {
	if middleware, exists := s.namedMiddlewares[name]; exists {
		return middleware()
	}
	// Route registration must fail closed when its protection is misspelled or
	// omitted from bootstrap. A no-op here would expose a protected endpoint.
	panic("middleware not registered: " + name)
}
