-- SELECT-only diagnostics, executed with default_transaction_read_only=on.
-- Output contains fixed check names and booleans, never identities or secrets.
SELECT 'business_migrations=' || CASE WHEN
  (SELECT count(*) = 1 AND bool_and(version = :'expected_version'::bigint AND NOT dirty) FROM public.schema_migrations)
  THEN 'ok' ELSE 'failed' END;
SELECT 'auth_migrations=' || CASE WHEN
  NOT EXISTS (
    SELECT 1 FROM auth.starter_auth_migrations applied
    FULL JOIN jsonb_to_recordset(:'expected_auth'::jsonb) AS expected(name text, checksum text) USING (name)
    WHERE applied.checksum IS DISTINCT FROM expected.checksum
  )
  THEN 'ok' ELSE 'failed' END;
SELECT 'database_roles=' || CASE WHEN
  (SELECT count(*) = 2 AND bool_and(rolcanlogin AND NOT rolsuper AND NOT rolbypassrls AND NOT rolcreatedb AND NOT rolcreaterole AND NOT rolinherit)
   FROM pg_roles WHERE rolname IN ('starter_app','starter_auth'))
  AND NOT EXISTS (
    SELECT 1 FROM pg_roles r WHERE (r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.oid IN (
      SELECT relowner FROM pg_class WHERE oid IN ('organizations.organizations'::regclass,'organizations.accounts'::regclass)
    )) AND pg_has_role('starter_app',r.oid,'MEMBER')
  ) THEN 'ok' ELSE 'failed' END;
SELECT 'tenant_rls=' || CASE WHEN
  (SELECT count(*) = 2 AND bool_and(relrowsecurity AND relforcerowsecurity) FROM pg_class
   WHERE oid IN ('organizations.organizations'::regclass,'organizations.accounts'::regclass))
  THEN 'ok' ELSE 'failed' END;
SELECT 'auth_isolation=' || CASE WHEN
  (SELECT nspowner = 'starter_auth'::regrole FROM pg_namespace WHERE nspname = 'auth')
  AND NOT has_schema_privilege('starter_app','auth','USAGE')
  AND NOT has_table_privilege('starter_app','auth.session','SELECT')
  THEN 'ok' ELSE 'failed' END;
SELECT 'membership_roles=' || CASE WHEN
  NOT EXISTS (SELECT 1 FROM auth.member WHERE role NOT IN ('admin','manager','member') OR role IS NULL)
  AND NOT EXISTS (SELECT 1 FROM auth.invitation WHERE role NOT IN ('admin','manager','member') OR role IS NULL)
  THEN 'ok' ELSE 'failed' END;
