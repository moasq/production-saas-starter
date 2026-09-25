-- Rollback quota_tracking schema changes
-- Restore invoice_count_max and invoice_count_current

-- Step 1: Add back old columns
ALTER TABLE subscription_billing.quota_tracking
ADD COLUMN invoice_count_current INT DEFAULT 0,
ADD COLUMN invoice_count_max INT;

-- Step 2: Migrate data back (this is a best-effort rollback, data may be lost)
-- We can't perfectly restore the original split, so we set current=0 and max=invoice_count
UPDATE subscription_billing.quota_tracking
SET invoice_count_current = 0,
    invoice_count_max = invoice_count;

-- Step 3: Make columns NOT NULL
ALTER TABLE subscription_billing.quota_tracking
ALTER COLUMN invoice_count_current SET NOT NULL,
ALTER COLUMN invoice_count_max SET NOT NULL;

-- Step 4: Drop new column
ALTER TABLE subscription_billing.quota_tracking
DROP COLUMN invoice_count;

-- Restore comments
COMMENT ON COLUMN subscription_billing.quota_tracking.invoice_count_current IS 'Current invoice count in billing period';
COMMENT ON COLUMN subscription_billing.quota_tracking.invoice_count_max IS 'Maximum invoices allowed from product metadata';
