package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const SchemaVersion = 11

// RuntimeQuerier is the read-only database boundary used at startup and readiness.
type RuntimeQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func ValidateRuntime(ctx context.Context, pool RuntimeQuerier) error {
	var elevated bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (
  SELECT 1 FROM pg_roles r WHERE
   (r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.oid IN (
    SELECT relowner FROM pg_class WHERE oid IN ('organizations.organizations'::regclass,'organizations.accounts'::regclass)
   )) AND pg_has_role(current_user,r.oid,'MEMBER')
 )`).Scan(&elevated); err != nil {
		return fmt.Errorf("check runtime role: %w", err)
	}
	if elevated {
		return fmt.Errorf("API database role must not own business tables, bypass RLS, create roles, or belong to an elevated role")
	}
	var version int
	var dirty bool
	if err := pool.QueryRow(ctx, "SELECT version,dirty FROM public.schema_migrations").Scan(&version, &dirty); err != nil {
		return fmt.Errorf("run api migrate before startup: %w", err)
	}
	if dirty || version != SchemaVersion {
		return fmt.Errorf("expected clean schema version %d; got version %d dirty=%t", SchemaVersion, version, dirty)
	}
	var enforced int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_class WHERE oid IN ('organizations.organizations'::regclass,'organizations.accounts'::regclass) AND relrowsecurity AND relforcerowsecurity`).Scan(&enforced); err != nil {
		return err
	}
	if enforced != 2 {
		return fmt.Errorf("tenant tables must enforce FORCE ROW LEVEL SECURITY")
	}
	return nil
}
