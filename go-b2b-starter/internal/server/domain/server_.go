package domain

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	config "github.com/moasq/go-b2b-starter/internal/server/config"
// 	"github.com/moasq/go-b2b-starter/internal/server/logging"
// 	"github.com/moasq/go-b2b-starter/internal/server/middleware"
// )

// // Constants
// const (
// 	ApiPrefix   = "/api"
// 	ApiVersion1 = "v1"
// )

// // Types and interfaces
// type RouteRegistrar func(*gin.RouterGroup)
// type MiddlewareFunc func() gin.HandlerFunc

// type Server interface {
// 	Start() error
// 	RegisterRoutes(registrar RouteRegistrar, prefix string, version ...string)
// 	RegisterNamedMiddleware(name string, middleware MiddlewareFunc)
// }

// type HTTPServer struct {
// 	config           *config.Config
// 	router           *gin.Engine
// 	logger           *logging.Logger
// 	securityLogger   *logging.SecurityLogger
// 	registrars       map[string][]RouteRegistrar
// 	namedMiddlewares map[string]MiddlewareFunc
// }

// // Constructor
// func NewHTTPServer(config *config.Config, router *gin.Engine, logger *logging.Logger) Server {
// 	if config.IsProd() {
// 		gin.SetMode(gin.ReleaseMode)
// 	}

// 	server := &HTTPServer{
// 		config:           config,
// 		router:           router,
// 		logger:           logger,
// 		securityLogger:   logging.NewSecurityLogger(logger.SugaredLogger),
// 		registrars:       make(map[string][]RouteRegistrar),
// 		namedMiddlewares: make(map[string]MiddlewareFunc),
// 	}

// 	server.setupMiddleware()
// 	return server
// }

// // Public methods
// func (s *HTTPServer) Start() error {
// 	srv := s.createHTTPServer()

// 	// Register all routes
// 	s.registerAllRoutes()

// 	// Setup health check
// 	s.setupHealthCheck()

// 	// Start server
// 	go s.startServer(srv)

// 	return s.handleGracefulShutdown(srv)
// }

// func (s *HTTPServer) RegisterRoutes(registrar RouteRegistrar, prefix string, version ...string) {
// 	v := ""
// 	if len(version) > 0 {
// 		v = version[0]
// 	}

// 	group := s.router.Group(prefix)
// 	if v != "" {
// 		group = group.Group("/" + v)
// 	}

// 	s.registrars[v] = append(s.registrars[v], func(g *gin.RouterGroup) {
// 		registrar(group)
// 	})
// }

// func (s *HTTPServer) RegisterNamedMiddleware(name string, middleware MiddlewareFunc) {
// 	s.namedMiddlewares[name] = middleware
// 	s.logger.Info("Named middleware registered: " + name)
// }

// // Private methods - Server setup and management
// func (s *HTTPServer) createHTTPServer() *http.Server {
// 	return &http.Server{
// 		Addr:              s.config.ServerAddress,
// 		Handler:           s.router,
// 		ReadTimeout:       15 * time.Second,
// 		WriteTimeout:      15 * time.Second,
// 		IdleTimeout:       60 * time.Second,
// 		ReadHeaderTimeout: 5 * time.Second,
// 		MaxHeaderBytes:    s.config.MaxRequestSize,
// 	}
// }

// func (s *HTTPServer) startServer(srv *http.Server) {
// 	s.logger.Info("Starting server on " + s.config.ServerAddress)
// 	var err error

// 	if s.config.IsProd() {
// 		err = srv.ListenAndServeTLS(
// 			s.config.TLSCertPath,
// 			s.config.TLSKeyPath,
// 		)
// 	} else {
// 		err = srv.ListenAndServe()
// 	}

// 	if err != nil && err != http.ErrServerClosed {
// 		s.logger.Fatal("Failed to start server", err)
// 	}
// }

// func (s *HTTPServer) handleGracefulShutdown(srv *http.Server) error {
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit
// 	s.logger.Info("Shutting down server...")

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if err := srv.Shutdown(ctx); err != nil {
// 		s.logger.Fatal("Server forced to shutdown", err)
// 	}

// 	s.logger.Info("Server exited gracefully")
// 	return nil
// }

// func (s *HTTPServer) registerAllRoutes() {
// 	for version, registrars := range s.registrars {
// 		group := s.router.Group("/api")
// 		if version != "" {
// 			group = group.Group("/" + version)
// 		}
// 		for _, registrar := range registrars {
// 			registrar(group)
// 		}
// 	}
// }

// // Private methods - Middleware setup
// func (s *HTTPServer) setupMiddleware() {
// 	ipProtection := middleware.NewIPProtection()

// 	// Add core middleware
// 	s.router.Use(
// 		middleware.SecurityHeaders(),
// 		ipProtection.Protect(),
// 		middleware.RequestSanitization(s.config.GetSanitizationConfig()),
// 		s.recoveryMiddleware(),
// 		middleware.RequestSizeLimit(int64(s.config.MaxRequestSize)),
// 		middleware.Timeout(10*time.Second),
// 		middleware.RateLimiter(s.config.RateLimitPerSecond),
// 		middleware.CORS(s.config.AllowedOrigins),
// 		s.requestLoggingMiddleware(),
// 	)

// 	if len(s.config.TrustedProxies) > 0 {
// 		s.router.SetTrustedProxies(s.config.TrustedProxies)
// 	}
// }

// func (s *HTTPServer) recoveryMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		defer func() {
// 			if err := recover(); err != nil {
// 				s.securityLogger.LogSecurityEvent(logging.SecurityEvent{
// 					EventType:   "PANIC_RECOVERED",
// 					IP:          c.ClientIP(),
// 					Description: fmt.Sprintf("Panic recovered: %v", err),
// 					Severity:    "HIGH",
// 					Timestamp:   time.Now(),
// 					RequestPath: c.Request.URL.Path,
// 					RequestID:   c.GetHeader("X-Request-ID"),
// 				})
// 				c.AbortWithStatus(http.StatusInternalServerError)
// 			}
// 		}()
// 		c.Next()
// 	}
// }

// func (s *HTTPServer) requestLoggingMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if s.config.IsProd() && c.Request.URL.Path == "/health" {
// 			c.Next()
// 			return
// 		}

// 		start := time.Now()
// 		path := c.Request.URL.Path
// 		query := c.Request.URL.RawQuery

// 		c.Next()

// 		s.logger.Infow("Request completed",
// 			"status", c.Writer.Status(),
// 			"method", c.Request.Method,
// 			"path", path,
// 			"query", query,
// 			"ip", c.ClientIP(),
// 			"latency", time.Since(start),
// 			"user-agent", c.Request.UserAgent(),
// 			"request-id", c.GetHeader("X-Request-ID"),
// 			"bytes-out", c.Writer.Size(),
// 		)
// 	}
// }

// // Private methods - Health check
// func (s *HTTPServer) setupHealthCheck() {
// 	s.router.GET("/health", func(c *gin.Context) {
// 		if s.config.IsProd() {
// 			c.JSON(http.StatusOK, gin.H{"status": "OK"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"status":      "OK",
// 			"environment": s.config.Env,
// 			"version":     "1.0.0",
// 			"timestamp":   time.Now().UTC(),
// 		})
// 	})
// 	s.logger.Info("Health check endpoint set up")
// }
