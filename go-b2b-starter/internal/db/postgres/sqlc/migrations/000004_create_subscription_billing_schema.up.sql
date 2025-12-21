-- Create subscription_billing schema for local subscription and quota tracking
CREATE SCHEMA IF NOT EXISTS subscription_billing;

-- Subscriptions table: Stores Polar subscription data locally for fast access
CREATE TABLE subscription_billing.subscriptions (
    id SERIAL PRIMARY KEY,
    organization_id INT NOT NULL UNIQUE REFERENCES organizations.organizations(id) ON DELETE CASCADE,

    -- Polar identifiers
    external_customer_id VARCHAR(100) NOT NULL,  -- Polar customer ID
    subscription_id VARCHAR(100) NOT NULL UNIQUE, -- Polar subscription ID

    -- Subscription details
    subscription_status VARCHAR(50) NOT NULL,     -- active, canceled, past_due, etc.
    product_id VARCHAR(100) NOT NULL,            -- Polar product ID
    product_name VARCHAR(255),                    -- Product display name
    plan_name VARCHAR(100),                       -- From subscription metadata

    -- Billing period
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,

    -- Cancellation details
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMP,

    -- Audit timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Metadata from Polar (stored as JSONB for flexibility)
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Quota tracking table: Tracks usage quotas per organization
CREATE TABLE subscription_billing.quota_tracking (
    id SERIAL PRIMARY KEY,
    organization_id INT NOT NULL UNIQUE REFERENCES organizations.organizations(id) ON DELETE CASCADE,

    -- Invoice quota (main quota we're tracking)
    invoice_count_current INT DEFAULT 0,         -- Current usage in period
    invoice_count_max INT NOT NULL,              -- Maximum allowed in period

    -- Additional quotas from product metadata
    max_seats INT,                                -- Maximum seats allowed

    -- Quota period (should match subscription period)
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,

    -- Sync tracking
    last_synced_at TIMESTAMP,                    -- Last time synced with Polar

    -- Audit timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast lookups
CREATE INDEX idx_subscriptions_organization_id ON subscription_billing.subscriptions(organization_id);
CREATE INDEX idx_subscriptions_subscription_status ON subscription_billing.subscriptions(subscription_status);
CREATE INDEX idx_subscriptions_external_customer_id ON subscription_billing.subscriptions(external_customer_id);

CREATE INDEX idx_quota_tracking_organization_id ON subscription_billing.quota_tracking(organization_id);
CREATE INDEX idx_quota_tracking_period_end ON subscription_billing.quota_tracking(period_end);

-- Comments for documentation
COMMENT ON SCHEMA subscription_billing IS 'Local cache of Polar subscription and quota data for fast access';
COMMENT ON TABLE subscription_billing.subscriptions IS 'Stores subscription details from Polar, synced via webhooks';
COMMENT ON TABLE subscription_billing.quota_tracking IS 'Tracks usage quotas per organization for fast quota checks';

COMMENT ON COLUMN subscription_billing.subscriptions.external_customer_id IS 'Polar customer ID (maps to organization via stytch_org_id)';
COMMENT ON COLUMN subscription_billing.quota_tracking.invoice_count_current IS 'Current invoice count in billing period';
COMMENT ON COLUMN subscription_billing.quota_tracking.invoice_count_max IS 'Maximum invoices allowed from product metadata';
