-- Update quota_tracking schema to use count-down system
-- Remove invoice_count_max, rename invoice_count_current to invoice_count (remaining invoices)

-- Step 1: Add new invoice_count column (temporary)
ALTER TABLE subscription_billing.quota_tracking
ADD COLUMN invoice_count INT;

-- Step 2: Migrate data: invoice_count = invoice_count_max - invoice_count_current
-- For existing rows, calculate remaining invoices
UPDATE subscription_billing.quota_tracking
SET invoice_count = GREATEST(invoice_count_max - invoice_count_current, 0);

-- Step 3: Make invoice_count NOT NULL with default 0
ALTER TABLE subscription_billing.quota_tracking
ALTER COLUMN invoice_count SET NOT NULL,
ALTER COLUMN invoice_count SET DEFAULT 0;

-- Step 4: Drop old columns
ALTER TABLE subscription_billing.quota_tracking
DROP COLUMN invoice_count_current,
DROP COLUMN invoice_count_max;

-- Update comment
COMMENT ON COLUMN subscription_billing.quota_tracking.invoice_count IS 'Remaining invoices in current billing period (decremented on use)';
