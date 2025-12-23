package domain

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) setupHealthCheck() {
	healthHandler := func(c *gin.Context) {
		if s.config.IsProd() {
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "OK",
			"environment": s.config.Env,
			"version":     "1.0.0",
			"timestamp":   time.Now().UTC(),
		})
	}

	// Register health endpoint at both paths
	s.router.GET("/health", healthHandler)
	s.router.GET("/api/health", healthHandler)
	s.logger.Info("Health check endpoints set up at /health and /api/health")
}

func (s *HTTPServer) setupRootEndpoint() {
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "B2B SaaS Starter API",
			"version":   "1.0.0",
			"status":    "running",
			"health":    "/api/health",
			"docs":      "/api/docs",
			"timestamp": time.Now().UTC(),
		})
	})
	s.logger.Info("Root endpoint set up at /")
}
