-- Rollback: Remove RBAC roles enum constraint

-- Drop the check constraint that enforces valid role values
ALTER TABLE organizations.accounts
DROP CONSTRAINT valid_role_enum;
