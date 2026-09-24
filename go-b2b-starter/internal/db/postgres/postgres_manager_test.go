package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
	"os"
	"testing"
	"time"
)

// TEST_DATABASE_URL must point to a disposable PostgreSQL server with CREATE
// DATABASE privileges. Each case gets its own database and removes only that DB.
func TestMigrationsFreshAndLegacy(t *testing.T) {
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	for _, version := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
		legacy := version > 0
		t.Run(fmt.Sprintf("legacy_%d", version), func(t *testing.T) {
			name := fmt.Sprintf("starter_test_%d", time.Now().UnixNano())
			identifier := pgx.Identifier{name}.Sanitize()
			if _, err := admin.Exec(ctx, "CREATE DATABASE "+identifier); err != nil {
				t.Fatal(err)
			}
			defer admin.Exec(ctx, "DROP DATABASE "+identifier+" WITH (FORCE)")
			parsed, err := url.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			parsed.Path = "/" + name
			pool, err := pgxpool.New(ctx, parsed.String())
			if err != nil {
				t.Fatal(err)
			}
			defer pool.Close()
			password, _ := parsed.User.Password()
			cfg := Config{Host: parsed.Hostname(), Port: parsed.Port(), User: parsed.User.Username(), Password: password, DBName: name, SSLMode: parsed.Query().Get("sslmode")}
			if cfg.Port == "" {
				cfg.Port = "5432"
			}
			if cfg.SSLMode == "" {
				cfg.SSLMode = "disable"
			}
			if legacy {
				if version >= 2 {
					sql, err := migrationFiles.ReadFile("archive/migrations/000002_create_organizations_schema.up.sql")
					if err != nil {
						t.Fatal(err)
					}
					if _, err := pool.Exec(ctx, string(sql)); err != nil {
						t.Fatal(err)
					}
					if _, err := pool.Exec(ctx, "INSERT INTO organizations.organizations(slug,name) VALUES('legacy-org','Existing company')"); err != nil {
						t.Fatal(err)
					}
				}
				if version >= 3 {
					sql, err := migrationFiles.ReadFile("archive/migrations/000003_enforce_role_enum.up.sql")
					if err != nil {
						t.Fatal(err)
					}
					if _, err := pool.Exec(ctx, string(sql)); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := pool.Exec(ctx, `CREATE TABLE schema_migrations(version bigint NOT NULL PRIMARY KEY,dirty boolean NOT NULL);
                    CREATE TABLE legacy_documents(id int PRIMARY KEY,body text NOT NULL);
                    INSERT INTO legacy_documents VALUES(1,'preserve existing data');`); err != nil {
					t.Fatal(err)
				}
				if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations VALUES($1,false)", version); err != nil {
					t.Fatal(err)
				}
			}

			manager := NewPostgresManager(cfg, pool)
			for i := 0; i < 2; i++ {
				if err := manager.RunMigrations(); err != nil {
					t.Fatal(err)
				}
			}
			var migratedVersion int
			var dirty bool
			if err := pool.QueryRow(ctx, "SELECT version,dirty FROM schema_migrations").Scan(&migratedVersion, &dirty); err != nil {
				t.Fatal(err)
			}
			if migratedVersion != 10 || dirty {
				t.Fatalf("version=%d dirty=%t", migratedVersion, dirty)
			}
			var orgID int
			if err := pool.QueryRow(ctx, "INSERT INTO organizations.organizations(slug,name) VALUES('test-company','Test company') RETURNING id").Scan(&orgID); err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, "INSERT INTO organizations.accounts(organization_id,email,full_name,role) VALUES($1,'manager@example.com','Manager','manager')", orgID); err != nil {
				t.Fatal(err)
			}
			if legacy {
				var data string
				if err := pool.QueryRow(ctx, "SELECT body FROM legacy_documents WHERE id=1").Scan(&data); err != nil || data != "preserve existing data" {
					t.Fatalf("legacy data changed: %q %v", data, err)
				}
				if version >= 2 {
					var n int
					if err := pool.QueryRow(ctx, "SELECT count(*) FROM organizations.organizations WHERE slug='legacy-org'").Scan(&n); err != nil || n != 1 {
						t.Fatalf("tenant lost: %d %v", n, err)
					}
				}
			} else {
				var extensions int
				if err := pool.QueryRow(ctx, "SELECT count(*) FROM pg_extension WHERE extname='vector'").Scan(&extensions); err != nil || extensions != 0 {
					t.Fatal("fresh database must not depend on vector")
				}
			}
			if _, err := pool.Exec(ctx, "UPDATE schema_migrations SET dirty=true"); err != nil {
				t.Fatal(err)
			}
			if err := manager.RunMigrations(); err == nil {
				t.Fatal("dirty migration state was ignored")
			}
		})
	}
}
