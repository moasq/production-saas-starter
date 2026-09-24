-- Automated downgrade intentionally refuses to destroy tenant data.
DO $$ BEGIN RAISE EXCEPTION 'Restore a verified backup to downgrade this schema'; END $$;
