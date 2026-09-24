-- Drop subscription billing schema and all its tables
DROP TABLE IF EXISTS subscription_billing.quota_tracking CASCADE;
DROP TABLE IF EXISTS subscription_billing.subscriptions CASCADE;
DROP SCHEMA IF EXISTS subscription_billing CASCADE;
