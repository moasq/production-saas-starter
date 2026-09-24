-- Enforce RBAC roles enum according to PERMISSIONS.md specification
-- Valid roles: member, approver, admin (plus legacy roles for backward compatibility)

-- Add check constraint to organizations.accounts.role column
-- Ensures only valid role values are stored in the database
ALTER TABLE organizations.accounts
ADD CONSTRAINT valid_role_enum CHECK (
    role IN ('member', 'approver', 'admin', 'owner', 'reviewer', 'employee')
);

-- Add comment documenting the role enum constraint
COMMENT ON CONSTRAINT valid_role_enum ON organizations.accounts IS
'RBAC Role Enum Constraint per PERMISSIONS.md:
- member: Process invoices day-to-day (Member role)
- approver: Review and approve invoices assigned to them (Approver role)
- admin: Full system control and management (Admin role)

Legacy roles (for backward compatibility during migration):
- owner: Legacy mapping to admin role
- reviewer: Legacy mapping to approver role
- employee: Legacy mapping to member role

All new role assignments MUST use the three core roles: member, approver, admin';
