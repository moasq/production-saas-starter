-- Drop triggers
DROP TRIGGER IF EXISTS trigger_accounts_updated_at ON organizations.accounts;
DROP TRIGGER IF EXISTS trigger_organizations_updated_at ON organizations.organizations;

-- Drop indexes (will be dropped automatically with tables, but being explicit)
DROP INDEX IF EXISTS organizations.idx_accounts_role;
DROP INDEX IF EXISTS organizations.idx_accounts_status;
DROP INDEX IF EXISTS organizations.idx_accounts_email;
DROP INDEX IF EXISTS organizations.idx_accounts_org_id;
DROP INDEX IF EXISTS organizations.idx_accounts_stytch_member_id;

DROP INDEX IF EXISTS organizations.idx_organizations_created_at;
DROP INDEX IF EXISTS organizations.idx_organizations_status;
DROP INDEX IF EXISTS organizations.idx_organizations_slug;
DROP INDEX IF EXISTS organizations.idx_organizations_stytch_org_id;

-- Drop tables in dependency order
DROP TABLE IF EXISTS organizations.accounts;
DROP TABLE IF EXISTS organizations.organizations;

-- Drop the schema
DROP SCHEMA IF EXISTS organizations CASCADE;
