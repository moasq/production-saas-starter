-- Auth cutover must be rolled back with the documented coordinated database backup.
-- Refuse a partial downgrade that would weaken tenant isolation or orphan identities.
DO $$ BEGIN
 RAISE EXCEPTION 'Restore the pre-cutover backup to roll back Better Auth and tenant RLS together';
END $$;
