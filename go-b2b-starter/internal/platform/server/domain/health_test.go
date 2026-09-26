package domain

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/moasq/go-b2b-starter/internal/db/postgres"
)

type healthDatabase struct {
	ping  func(context.Context) error
	query func(context.Context, string, ...any) pgx.Row
}

func (db healthDatabase) Ping(ctx context.Context) error { return db.ping(ctx) }
func (db healthDatabase) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.query(ctx, sql, args...)
}

type healthRow func(...any) error

func (r healthRow) Scan(dest ...any) error { return r(dest...) }

func healthRouter(db healthDatabase) *gin.Engine {
	router := gin.New()
	server := &HTTPServer{router: router, database: db, logger: &capturedLogger{}, runtimeReadiness: func(ctx context.Context) error { return postgres.ValidateRuntime(ctx, db) }}
	server.setupHealthCheck()
	return router
}
func healthyDatabase() healthDatabase {
	return healthDatabase{
		ping: func(context.Context) error { return nil },
		query: func(_ context.Context, sql string, _ ...any) pgx.Row {
			return healthRow(func(dest ...any) error {
				switch {
				case strings.Contains(sql, "pg_roles"):
					*dest[0].(*bool) = false
				case strings.Contains(sql, "schema_migrations"):
					*dest[0].(*int) = postgres.SchemaVersion
					*dest[1].(*bool) = false
				default:
					*dest[0].(*int) = 2
				}
				return nil
			})
		},
	}
}
func TestLivenessNeverTouchesDependencies(t *testing.T) {
	db := healthDatabase{
		ping:  func(context.Context) error { t.Fatal("liveness touched database"); return nil },
		query: func(context.Context, string, ...any) pgx.Row { t.Fatal("liveness touched schema"); return nil },
	}
	for _, path := range []string{"/livez", "/api/livez"} {
		response := httptest.NewRecorder()
		healthRouter(db).ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || response.Body.String() != `{"status":"alive"}` {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body)
		}
		if response.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("health must not be cached")
		}
	}
}
func TestReadinessChecksDatabaseContractAndKeepsLegacyRoutes(t *testing.T) {
	for _, path := range []string{"/readyz", "/api/readyz", "/health", "/api/health"} {
		response := httptest.NewRecorder()
		healthRouter(healthyDatabase()).ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"database_contract":"ok"`) {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body)
		}
	}
}
func TestReadinessFailuresExposeOnlyFixedCheckNames(t *testing.T) {
	for _, failure := range []string{"database", "role", "migration_dirty", "migration_version", "rls", "query"} {
		t.Run(failure, func(t *testing.T) {
			db := healthyDatabase()
			original := db.query
			if failure == "database" {
				db.ping = func(context.Context) error { return errors.New("postgres://private:secret@host/db") }
			} else {
				db.query = func(ctx context.Context, sql string, args ...any) pgx.Row {
					return healthRow(func(dest ...any) error {
						if failure == "query" {
							return errors.New("private database error with credentials")
						}
						if err := original(ctx, sql, args...).Scan(dest...); err != nil {
							return err
						}
						switch {
						case failure == "role" && strings.Contains(sql, "pg_roles"):
							*dest[0].(*bool) = true
						case failure == "migration_dirty" && strings.Contains(sql, "schema_migrations"):
							*dest[1].(*bool) = true
						case failure == "migration_version" && strings.Contains(sql, "schema_migrations"):
							*dest[0].(*int) = postgres.SchemaVersion - 1
						case failure == "rls" && strings.Contains(sql, "count(*)"):
							*dest[0].(*int) = 1
						}
						return nil
					})
				}
			}
			response := httptest.NewRecorder()
			healthRouter(db).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			expected := `{"checks":{"database":"ok","database_contract":"failed"},"status":"not_ready"}`
			if failure == "database" {
				expected = `{"checks":{"database":"failed","database_contract":"not_checked"},"status":"not_ready"}`
			}
			if response.Code != http.StatusServiceUnavailable || response.Body.String() != expected {
				t.Fatalf("%d %s", response.Code, response.Body)
			}
		})
	}
}
func TestReadinessHasOneBoundedDeadlineForPingAndContract(t *testing.T) {
	for _, stage := range []string{"ping", "contract"} {
		t.Run(stage, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				db := healthyDatabase()
				db.ping = func(ctx context.Context) error {
					if stage == "ping" {
						<-ctx.Done()
						return ctx.Err()
					}
					time.Sleep(time.Second)
					return nil
				}
				db.query = func(ctx context.Context, _ string, _ ...any) pgx.Row {
					return healthRow(func(...any) error { <-ctx.Done(); return ctx.Err() })
				}
				start := time.Now()
				response := httptest.NewRecorder()
				healthRouter(db).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
				if response.Code != http.StatusServiceUnavailable || time.Since(start) != readinessTimeout {
					t.Fatalf("status=%d duration=%s", response.Code, time.Since(start))
				}
			})
		})
	}
}
func TestReadinessHonorsClientCancellation(t *testing.T) {
	db := healthyDatabase()
	db.ping = func(ctx context.Context) error { return ctx.Err() }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil).WithContext(ctx)
	healthRouter(db).ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("cancelled request: %d", response.Code)
	}
}
