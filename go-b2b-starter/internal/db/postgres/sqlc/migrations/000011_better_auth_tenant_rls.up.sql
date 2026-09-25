-- Keep historical provider columns for reversible identity cutover and billing linkage.
ALTER TABLE organizations.organizations ADD COLUMN auth_org_id TEXT;
ALTER TABLE organizations.organizations ADD COLUMN polar_customer_external_id TEXT;
UPDATE organizations.organizations SET auth_org_id = COALESCE(NULLIF(stytch_org_id,''),'legacy-org-' || id),
 polar_customer_external_id = COALESCE(NULLIF(stytch_org_id,''),'legacy-org-' || id);
ALTER TABLE organizations.organizations ALTER COLUMN auth_org_id SET NOT NULL;
ALTER TABLE organizations.organizations ALTER COLUMN polar_customer_external_id SET NOT NULL;
ALTER TABLE organizations.organizations ADD CONSTRAINT organizations_auth_org_id_key UNIQUE(auth_org_id);
-- Resolve case-only duplicate memberships before cutover; never merge tenants silently.
CREATE UNIQUE INDEX accounts_tenant_normalized_email_key ON organizations.accounts(organization_id,lower(email));
-- Preserve per-tenant profile evidence before memberships share a Better Auth user.
-- Snapshot every pre-cutover account, including older rows without provider IDs.
-- New accounts leave this nullable column empty; normal synchronization never writes it.
ALTER TABLE organizations.accounts ADD COLUMN legacy_full_name TEXT;
UPDATE organizations.accounts SET legacy_full_name=full_name;
ALTER TABLE organizations.accounts ADD COLUMN auth_user_id TEXT;
ALTER TABLE organizations.accounts ADD COLUMN auth_member_id TEXT;
ALTER TABLE organizations.accounts ADD CONSTRAINT accounts_auth_member_id_key UNIQUE(auth_member_id);

ALTER TABLE organizations.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations.organizations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_organization ON organizations.organizations
 USING (auth_org_id = NULLIF(current_setting('app.tenant_id',true),''))
 WITH CHECK (auth_org_id = NULLIF(current_setting('app.tenant_id',true),''));
ALTER TABLE organizations.accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations.accounts FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_account ON organizations.accounts
 USING (organization_id IN (SELECT id FROM organizations.organizations))
 WITH CHECK (organization_id IN (SELECT id FROM organizations.organizations));

-- The runtime role is provisioned separately and never owns business tables.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='starter_app') THEN
  GRANT USAGE ON SCHEMA organizations TO starter_app;
  GRANT SELECT,INSERT,UPDATE,DELETE ON organizations.organizations,organizations.accounts TO starter_app;
  GRANT USAGE,SELECT ON SEQUENCE organizations.organizations_id_seq,organizations.accounts_id_seq TO starter_app;
  GRANT SELECT ON public.schema_migrations TO starter_app;
 END IF;
END $$;
