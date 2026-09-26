package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessTimeout = 2 * time.Second

func (s *HTTPServer) setupHealthCheck() {
	// Liveness only asks whether this process can serve HTTP. Database, SMTP,
	// auth bridge and billing outages must not cause a liveness restart loop.
	live := func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	}
	ready := func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
		defer cancel()
		checks := gin.H{"database": "failed", "database_contract": "not_checked"}
		if s.database == nil || s.database.Ping(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "checks": checks})
			return
		}
		checks["database"] = "ok"
		checks["database_contract"] = "failed"
		// Recheck the startup contract: clean expected migration, a restricted
		// runtime role, and FORCE RLS. SELECT-only and bounded by the same deadline.
		if s.runtimeReadiness == nil || s.runtimeReadiness(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "checks": checks})
			return
		}
		checks["database_contract"] = "ok"
		c.JSON(http.StatusOK, gin.H{"status": "ready", "checks": checks})
	}
	for _, path := range []string{"/livez", "/api/livez"} {
		s.router.GET(path, live)
	}
	// Preserve /health for existing deployments; explicitly use /livez for
	// process restarts and /readyz for traffic admission/startup dependencies.
	for _, path := range []string{"/readyz", "/api/readyz", "/health", "/api/health"} {
		s.router.GET(path, ready)
	}
	s.logger.Info("Liveness at /livez; database readiness at /readyz (legacy /health)")
}

func (s *HTTPServer) setupRootEndpoint() {
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "B2B SaaS Starter API",
			"version":   "1.0.0",
			"status":    "running",
			"health":    "/api/health",
			"liveness":  "/livez",
			"readiness": "/readyz",
			"timestamp": time.Now().UTC(),
		})
	})
	s.logger.Info("Root endpoint set up at /")
}
