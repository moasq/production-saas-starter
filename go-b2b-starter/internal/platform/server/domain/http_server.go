package domain

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/moasq/go-b2b-starter/internal/platform/server/config"
	"github.com/moasq/go-b2b-starter/internal/platform/server/logging"
	"github.com/moasq/go-b2b-starter/internal/platform/server/middleware"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	config           *config.Config
	router           *gin.Engine
	logger           *logging.Logger
	securityLogger   *logging.SecurityLogger
	registrars       map[string][]RouteRegistrar
	namedMiddlewares map[string]MiddlewareFunc
	ipProtection     *middleware.IPProtection
}

func NewHTTPServer(
	config *config.Config,
	router *gin.Engine,
	logger *logging.Logger,
) Server {
	if config.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	ipProtection := middleware.NewIPProtection()

	server := &HTTPServer{
		config:           config,
		router:           router,
		logger:           logger,
		securityLogger:   logging.NewSecurityLogger(logger.SugaredLogger),
		registrars:       make(map[string][]RouteRegistrar),
		namedMiddlewares: make(map[string]MiddlewareFunc),
		ipProtection:     ipProtection,
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

func (s *HTTPServer) MiddlewareResolver() MiddlewareResolver {
	return s
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
		WriteTimeout:      30 * time.Second, // Increased to accommodate auto-extraction processing
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    s.config.MaxRequestSize,
	}
}

func (s *HTTPServer) startServer(srv *http.Server) {
	s.logger.Info("Starting server on " + s.config.ServerAddress)
	var err error

	if s.config.IsProd() {
		err = srv.ListenAndServeTLS(
			s.config.TLSCertPath,
			s.config.TLSKeyPath,
		)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("Failed to start server", err)
	}
}

func (s *HTTPServer) handleGracefulShutdown(srv *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("Shutting down server...")

	// Stop IP Protection cleanup goroutine
	if s.ipProtection != nil {
		s.ipProtection.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Fatal("Server forced to shutdown", err)
	}

	s.logger.Info("Server exited gracefully")
	return nil
}

// Get implements the MiddlewareResolver interface
func (s *HTTPServer) Get(name string) gin.HandlerFunc {
	if middleware, exists := s.namedMiddlewares[name]; exists {
		return middleware()
	}
	// Return a no-op middleware if not found
	return func(c *gin.Context) {
		s.logger.Warnw("Middleware not found", "name", name)
		c.Next()
	}
}

// GetMiddleware returns a middleware by name (compatibility method)
func (s *HTTPServer) GetMiddleware(name string) gin.HandlerFunc {
	return s.Get(name)
}
