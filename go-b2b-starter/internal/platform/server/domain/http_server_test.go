package domain

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"github.com/moasq/go-b2b-starter/internal/platform/server/config"
)

type capturedLogger struct{ entries []logger.Fields }

func (l *capturedLogger) Debug(string, ...logger.Fields) {}
func (l *capturedLogger) Info(_ string, fields ...logger.Fields) {
	l.entries = append(l.entries, fields...)
}
func (l *capturedLogger) Warn(string, ...logger.Fields) {}
func (l *capturedLogger) Error(_ string, fields ...logger.Fields) {
	l.entries = append(l.entries, fields...)
}
func (l *capturedLogger) Fatal(string, ...logger.Fields)         { panic("unexpected fatal log") }
func (l *capturedLogger) WithFields(logger.Fields) logger.Logger { return l }

func testServer() (*HTTPServer, *gin.Engine, *capturedLogger) {
	router := gin.New()
	log := &capturedLogger{}
	srv := NewHTTPServer(&config.Config{
		Env:                config.DEV,
		AllowedOrigins:     []string{"http://localhost:3000"},
		MaxRequestSize:     1024,
		RateLimitPerSecond: 100,
	}, router, log, nil).(*HTTPServer)
	return srv, router, log
}

func TestUnregisteredMiddlewareStopsRouteRegistration(t *testing.T) {
	srv, _, _ := testServer()
	defer func() {
		if recovered := recover(); recovered != "middleware not registered: auth" {
			t.Fatalf("expected missing protection to stop startup, got %v", recovered)
		}
	}()
	srv.RegisterRoutes(func(group *gin.RouterGroup, resolver MiddlewareResolver) {
		group.GET("/private", resolver.Get("auth"), func(c *gin.Context) { c.Status(http.StatusOK) })
	}, "/api")
}

func TestRegisteredMiddlewareStillProtectsRoutes(t *testing.T) {
	srv, router, _ := testServer()
	called := false
	srv.RegisterNamedMiddleware("auth", func() gin.HandlerFunc {
		return func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
	})
	srv.RegisterRoutes(func(group *gin.RouterGroup, resolver MiddlewareResolver) {
		group.GET("/private", resolver.Get("auth"), func(c *gin.Context) { called = true; c.Status(http.StatusOK) })
	}, "/api")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/private", nil))
	if response.Code != http.StatusUnauthorized || called {
		t.Fatalf("protected handler executed: status=%d called=%v", response.Code, called)
	}
}

func TestQueryValuesRemainApplicationDataAndAreNotLogged(t *testing.T) {
	_, router, log := testServer()
	router.GET("/search", func(c *gin.Context) { c.String(http.StatusOK, c.Query("name")) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/search?name=Select+Union&token=private-value", nil))
	if response.Code != http.StatusOK || response.Body.String() != "Select Union" {
		t.Fatalf("ordinary query data rejected: %d %s", response.Code, response.Body)
	}
	assertNoCredentialLogs(t, log)
}

func TestRecoveryReturnsSafeErrorAndDoesNotLogCredentials(t *testing.T) {
	_, router, log := testServer()
	router.GET("/panic", func(*gin.Context) { panic("handler failed") })
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic?token=private-value", nil)
	request.Header.Set("Authorization", "Bearer private-value")
	router.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "handler failed") {
		t.Fatalf("unsafe recovery response: %d %s", response.Code, response.Body)
	}
	var body struct {
		RequestID string `json:"request_id"`
		Code      string `json:"code"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.RequestID == "" || body.Code != "SERVER_ERROR" {
		t.Fatalf("missing safe error details: %s", response.Body)
	}
	assertNoCredentialLogs(t, log)
}

func assertNoCredentialLogs(t *testing.T, log *capturedLogger) {
	t.Helper()
	if len(log.entries) == 0 {
		t.Fatal("expected structured request or recovery logs")
	}
	data, err := json.Marshal(log.entries)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-value") {
		t.Fatalf("credentials leaked in logs: %s", data)
	}
}
