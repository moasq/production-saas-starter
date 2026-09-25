package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/db/tenant"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations"
)

func TestTenantRLSIsolationAndLegacyBillingMapping(t *testing.T) {
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL server")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	suffix := fmt.Sprint(time.Now().UnixNano())
	dbName := "rls_test_" + suffix
	roleName := "rls_app_" + suffix
	mustExec := func(pool *pgxpool.Pool, sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(admin, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize())

	mustExec(admin, "CREATE ROLE "+pgx.Identifier{roleName}.Sanitize()+" LOGIN NOSUPERUSER NOBYPASSRLS PASSWORD 'disposable-test-password'")
	defer func() {
		admin.Exec(ctx, "DROP DATABASE "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
		admin.Exec(ctx, "DROP ROLE "+pgx.Identifier{roleName}.Sanitize())
	}()
	parsed, _ := url.Parse(raw)
	parsed.Path = "/" + dbName
	owner, err := pgxpool.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	baseline, err := migrationFiles.ReadFile("sqlc/migrations/000010_minimal_b2b.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	mustExec(owner, string(baseline))
	mustExec(owner, "INSERT INTO organizations.organizations(slug,name,stytch_org_id) VALUES('legacy-company','Legacy company','organization-test-legacy')")
	mustExec(owner, "INSERT INTO organizations.accounts(organization_id,email,full_name,role,stytch_member_id) VALUES(1,'old@example.com','Original Name','admin','member-test-legacy')")
	mustExec(owner, "INSERT INTO organizations.organizations(slug,name,stytch_org_id) VALUES('second-company','Second company','organization-test-second')")
	mustExec(owner, "INSERT INTO organizations.accounts(organization_id,email,full_name,role,stytch_member_id) VALUES(2,'old@example.com','Second Original Name','member','member-test-second')")
	mustExec(owner, "CREATE TABLE schema_migrations(version bigint NOT NULL PRIMARY KEY,dirty boolean NOT NULL); INSERT INTO schema_migrations VALUES(10,false)")
	password, _ := parsed.User.Password()
	port := parsed.Port()
	if port == "" {
		port = "5432"
	}
	ssl := parsed.Query().Get("sslmode")
	if ssl == "" {
		ssl = "disable"
	}
	cfg := Config{Host: parsed.Hostname(), Port: port, User: parsed.User.Username(), Password: password, DBName: dbName, SSLMode: ssl}
	manager := NewPostgresManager(cfg, owner)
	if err := manager.RunMigrations(); err != nil {
		t.Fatal(err)
	}
	if err := manager.RunMigrations(); err != nil {
		t.Fatal(err)
	}
	mustExec(owner, "GRANT USAGE ON SCHEMA organizations TO "+pgx.Identifier{roleName}.Sanitize())
	mustExec(owner, "GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA organizations TO "+pgx.Identifier{roleName}.Sanitize())
	mustExec(owner, "GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA organizations TO "+pgx.Identifier{roleName}.Sanitize())
	mustExec(owner, "GRANT SELECT ON schema_migrations TO "+pgx.Identifier{roleName}.Sanitize())
	appURL := *parsed
	appURL.User = url.UserPassword(roleName, "disposable-test-password")
	poolCfg, err := pgxpool.ParseConfig(appURL.String())
	if err != nil {
		t.Fatal(err)
	}
	poolCfg.MaxConns = 1
	app, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	if err := ValidateRuntime(ctx, app); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntime(ctx, owner); err == nil {
		t.Fatal("privileged runtime credential accepted")
	}
	elevatedRole := "rls_bypass_" + suffix
	mustExec(admin, "CREATE ROLE "+pgx.Identifier{elevatedRole}.Sanitize()+" NOLOGIN BYPASSRLS")
	mustExec(admin, "GRANT "+pgx.Identifier{elevatedRole}.Sanitize()+" TO "+pgx.Identifier{roleName}.Sanitize())
	if err := ValidateRuntime(ctx, app); err == nil {
		t.Fatal("membership in bypass role accepted")
	}
	mustExec(admin, "REVOKE "+pgx.Identifier{elevatedRole}.Sanitize()+" FROM "+pgx.Identifier{roleName}.Sanitize())
	mustExec(admin, "DROP ROLE "+pgx.Identifier{elevatedRole}.Sanitize())
	store := tenant.New(app)
	if err := store.Run(ctx, func(*sqlc.Queries) error { return nil }); err == nil {
		t.Fatal("missing scope accepted")
	}
	count := func(want int) {
		t.Helper()
		var n int
		if err := app.QueryRow(ctx, "SELECT count(*) FROM organizations.organizations").Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != want {
			t.Fatalf("unscoped visible rows=%d want=%d", n, want)
		}
	}
	count(0)
	if _, err := app.Exec(ctx, "INSERT INTO organizations.organizations(name,slug,auth_org_id,polar_customer_external_id) VALUES('Leak','leak','leak','leak')"); err == nil {
		t.Fatal("unscoped write accepted")
	}
	var ids [2]int32
	for n, authID := range []string{"organization-test-legacy", "new-organization"} {
		scoped := tenant.WithScope(ctx, authID)
		err := store.Run(scoped, func(q *sqlc.Queries) error {
			org, err := q.SyncOrganization(scoped, sqlc.SyncOrganizationParams{AuthOrgID: authID, Name: fmt.Sprintf("Company %d", n), Slug: fmt.Sprintf("company-%d", n)})
			if err != nil {
				return err
			}
			ids[n] = org.ID
			_, err = q.SyncAccount(scoped, sqlc.SyncAccountParams{OrganizationID: org.ID, Email: fmt.Sprintf("user%d@example.com", n), FullName: "User", Role: "admin", AuthUserID: pgtype.Text{String: fmt.Sprintf("user-%d", n), Valid: true}, AuthMemberID: pgtype.Text{String: fmt.Sprintf("member-%d", n), Valid: true}})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		count(0) // same pool connection must lose SET LOCAL after every commit.
	}
	first := tenant.WithScope(ctx, "organization-test-legacy")
	if err := store.Run(first, func(q *sqlc.Queries) error {
		org, err := q.GetTenantOrganization(first, ids[0])
		if err != nil {
			return err
		}
		if org.PolarCustomerExternalID != "organization-test-legacy" || org.StytchOrgID.String != "organization-test-legacy" {
			t.Fatal("legacy billing identity changed")
		}
		if _, err := q.GetTenantOrganization(first, ids[1]); err != pgx.ErrNoRows {
			t.Fatalf("cross-tenant read returned %v", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Run(first, func(q *sqlc.Queries) error {
		_, err := q.SyncAccount(first, sqlc.SyncAccountParams{OrganizationID: ids[1], Email: "attack@example.com", FullName: "Attack", Role: "admin"})
		return err
	}); err == nil {
		t.Fatal("cross-tenant insert accepted")
	}
	count(0) // scope also resets after failed transactions.
	// Verified email case changes must preserve the original tenant account ID.
	if err := store.Run(first, func(q *sqlc.Queries) error {
		account, err := q.SyncAccount(first, sqlc.SyncAccountParams{OrganizationID: ids[0], Email: "OLD@EXAMPLE.COM", FullName: "Shared Better Auth Name", Role: "admin", AuthUserID: pgtype.Text{String: "legacy-user", Valid: true}, AuthMemberID: pgtype.Text{String: "legacy-member", Valid: true}})
		if err != nil {
			return err
		}
		if account.ID != 1 || account.Email != "old@example.com" || account.LegacyFullName.String != "Original Name" || account.FullName != "Shared Better Auth Name" {
			t.Fatal("normalized email created a duplicate account")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	secondLegacy := tenant.WithScope(ctx, "organization-test-second")
	if err := store.Run(secondLegacy, func(q *sqlc.Queries) error {
		account, err := q.SyncAccount(secondLegacy, sqlc.SyncAccountParams{OrganizationID: 2, Email: "old@example.com", FullName: "Shared Better Auth Name", Role: "member", AuthUserID: pgtype.Text{String: "legacy-user", Valid: true}, AuthMemberID: pgtype.Text{String: "legacy-member-second", Valid: true}})
		if err != nil {
			return err
		}
		if account.LegacyFullName.String != "Second Original Name" || account.FullName != "Shared Better Auth Name" {
			t.Fatal("shared identity erased distinct tenant profile evidence")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// SQL which forgets a tenant predicate is still restricted by PostgreSQL.
	tx, err := app.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id',$1,true)", "organization-test-legacy"); err != nil {
		t.Fatal(err)
	}
	tag, err := tx.Exec(ctx, "UPDATE organizations.accounts SET full_name='Changed' WHERE organization_id=$1", ids[1])
	if err != nil || tag.RowsAffected() != 0 {
		t.Fatalf("cross-tenant update: %d %v", tag.RowsAffected(), err)
	}
	tag, err = tx.Exec(ctx, "DELETE FROM organizations.organizations WHERE id=$1", ids[1])
	if err != nil || tag.RowsAffected() != 0 {
		t.Fatalf("cross-tenant delete: %d %v", tag.RowsAffected(), err)
	}
	var n int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM organizations.accounts").Scan(&n); err != nil || n != 2 {
		t.Fatalf("tenant read sees %d accounts: %v", n, err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	count(0)
	// A live identity cannot reactivate a suspended business tenant.
	mustExec(owner, "UPDATE organizations.organizations SET status='suspended' WHERE id=$1", ids[0])
	identity := &auth.Identity{OrganizationID: "organization-test-legacy", UserID: "user-0", MemberID: "member-0", Email: "user0@example.com", Roles: []auth.Role{auth.RoleAdmin}}
	identity.Organization.Name = "Renamed"
	identity.Organization.Slug = "renamed"
	identity.User.Name = "User"
	if _, _, err := organizations.NewResolver(store).Resolve(ctx, identity); err == nil {
		t.Fatal("suspended tenant reactivated")
	}
	var status string
	if err := owner.QueryRow(ctx, "SELECT status FROM organizations.organizations WHERE id=$1", ids[0]).Scan(&status); err != nil || status != "suspended" {
		t.Fatal("suspended tenant state overwritten")
	}
	// Original profile evidence survives first login even after adopting the global name.
	var name, memberID string
	if err := owner.QueryRow(ctx, "SELECT legacy_full_name,stytch_member_id FROM organizations.accounts WHERE organization_id=1 AND email='old@example.com'").Scan(&name, &memberID); err != nil || name != "Original Name" || memberID != "member-test-legacy" {
		t.Fatalf("legacy account changed: %s %s %v", name, memberID, err)
	}
}
