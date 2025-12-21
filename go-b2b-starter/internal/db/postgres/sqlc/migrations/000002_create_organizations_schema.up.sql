-- Create organizations schema
CREATE SCHEMA IF NOT EXISTS organizations;

-- Organizations table (top-level tenant)
CREATE TABLE organizations.organizations (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,

    -- Stytch linkage
    stytch_org_id VARCHAR(100) UNIQUE,
    stytch_connection_id VARCHAR(100),
    stytch_connection_name VARCHAR(255),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT chk_organizations_status CHECK (status IN ('active', 'suspended', 'cancelled')),
    CONSTRAINT chk_organizations_slug CHECK (slug ~ '^[a-z0-9-]+$' AND LENGTH(slug) >= 3)
);

-- Accounts table (users within organizations)
CREATE TABLE organizations.accounts (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,

    -- Account info
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    stytch_member_id VARCHAR(100),
    stytch_role_id VARCHAR(100),
    stytch_role_slug VARCHAR(100),
    stytch_email_verified BOOLEAN DEFAULT FALSE NOT NULL,

    -- Role and status (legacy field retained for business logic)
    role VARCHAR(50) DEFAULT 'member' NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,

    -- Activity tracking
    last_login_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Unique constraints
    UNIQUE(organization_id, email),
    UNIQUE(organization_id, stytch_member_id),

    -- Check constraints
    CONSTRAINT chk_accounts_role CHECK (role IN ('owner', 'admin', 'member', 'reviewer', 'employee')),
    CONSTRAINT chk_accounts_status CHECK (status IN ('active', 'inactive', 'suspended')),
    CONSTRAINT chk_accounts_email CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Indexes for performance
CREATE INDEX idx_organizations_slug ON organizations.organizations(slug) WHERE status = 'active';
CREATE INDEX idx_organizations_status ON organizations.organizations(status);
CREATE INDEX idx_organizations_created_at ON organizations.organizations(created_at DESC);
CREATE UNIQUE INDEX idx_organizations_stytch_org_id ON organizations.organizations(stytch_org_id);

CREATE INDEX idx_accounts_org_id ON organizations.accounts(organization_id);
CREATE INDEX idx_accounts_email ON organizations.accounts(email);
CREATE INDEX idx_accounts_status ON organizations.accounts(status);
CREATE INDEX idx_accounts_role ON organizations.accounts(role);
CREATE INDEX idx_accounts_stytch_member_id ON organizations.accounts(stytch_member_id);

-- Updated at trigger function (reuse existing function if available)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to automatically update updated_at
CREATE TRIGGER trigger_organizations_updated_at
    BEFORE UPDATE ON organizations.organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_accounts_updated_at
    BEFORE UPDATE ON organizations.accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON SCHEMA organizations IS 'Schema for organization and account management';
COMMENT ON TABLE organizations.organizations IS 'Organizations (tenants) in the system';
COMMENT ON TABLE organizations.accounts IS 'User accounts within organizations';
COMMENT ON COLUMN organizations.organizations.slug IS 'URL-friendly unique identifier for organization';
COMMENT ON COLUMN organizations.organizations.stytch_org_id IS 'Stytch organization identifier (org_xxx)';
COMMENT ON COLUMN organizations.organizations.stytch_connection_id IS 'Optional Stytch connection or project identifier associated with the organization';
COMMENT ON COLUMN organizations.organizations.stytch_connection_name IS 'Optional Stytch connection name associated with the organization';
COMMENT ON COLUMN organizations.accounts.stytch_member_id IS 'Stytch member identifier (member_xxx)';
COMMENT ON COLUMN organizations.accounts.stytch_role_id IS 'Stytch role identifier assigned to the member';
COMMENT ON COLUMN organizations.accounts.stytch_role_slug IS 'Human-readable Stytch role slug assigned to the member';
COMMENT ON COLUMN organizations.accounts.stytch_email_verified IS 'Whether Stytch reports the member email as verified';
COMMENT ON COLUMN organizations.accounts.role IS 'Last known role for business logic (e.g., owner, reviewer, employee)';
