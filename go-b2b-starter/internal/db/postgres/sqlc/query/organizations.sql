-- name: CreateOrganization :one
INSERT INTO organizations.organizations (
    slug,
    name,
    status
) VALUES (
    $1,
    $2,
    $3
) RETURNING
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at;

-- name: GetOrganizationByID :one
SELECT
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at
FROM organizations.organizations
WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at
FROM organizations.organizations
WHERE slug = $1;

-- name: GetOrganizationByStytchID :one
SELECT
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at
FROM organizations.organizations
WHERE stytch_org_id = $1;

-- name: UpdateOrganization :one
UPDATE organizations.organizations
SET
    name = $2,
    status = $3,
    stytch_org_id = $4,
    stytch_connection_id = $5,
    stytch_connection_name = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at;

-- name: UpdateOrganizationStytchInfo :one
UPDATE organizations.organizations
SET
    stytch_org_id = $2,
    stytch_connection_id = $3,
    stytch_connection_name = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at;

-- name: ListOrganizations :many
SELECT
    id,
    slug,
    name,
    status,
    stytch_org_id,
    stytch_connection_id,
    stytch_connection_name,
    created_at,
    updated_at
FROM organizations.organizations
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteOrganization :exec
DELETE FROM organizations.organizations
WHERE id = $1;

-- Accounts queries

-- name: CreateAccount :one
INSERT INTO organizations.accounts (
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
) RETURNING
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at;

-- name: GetAccountByID :one
SELECT
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at
FROM organizations.accounts
WHERE id = $1 AND organization_id = $2;

-- name: GetAccountByEmail :one
SELECT
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at
FROM organizations.accounts
WHERE email = $1 AND organization_id = $2;

-- name: ListAccountsByOrganization :many
SELECT
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at
FROM organizations.accounts
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: UpdateAccount :one
UPDATE organizations.accounts
SET
    full_name = $3,
    stytch_role_id = $4,
    stytch_role_slug = $5,
    stytch_email_verified = $6,
    role = $7,
    status = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2
RETURNING
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at;

-- name: UpdateAccountStytchInfo :one
UPDATE organizations.accounts
SET
    stytch_member_id = $3,
    stytch_role_id = $4,
    stytch_role_slug = $5,
    stytch_email_verified = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2
RETURNING
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at;

-- name: UpdateAccountLastLogin :one
UPDATE organizations.accounts
SET
    last_login_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2
RETURNING
    id,
    organization_id,
    email,
    full_name,
    stytch_member_id,
    stytch_role_id,
    stytch_role_slug,
    stytch_email_verified,
    role,
    status,
    last_login_at,
    created_at,
    updated_at;

-- name: DeleteAccount :exec
UPDATE organizations.accounts
SET
    status = 'inactive',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2;

-- Organization membership queries

-- name: GetOrganizationByUserEmail :one
SELECT
    o.id,
    o.slug,
    o.name,
    o.status,
    o.stytch_org_id,
    o.stytch_connection_id,
    o.stytch_connection_name,
    o.created_at,
    o.updated_at
FROM organizations.organizations o
INNER JOIN organizations.accounts a ON o.id = a.organization_id
WHERE a.email = $1
  AND a.status = 'active'
  AND o.status = 'active'
LIMIT 1;

-- name: GetAccountOrganization :one
SELECT
    o.id,
    o.slug,
    o.name,
    o.status,
    o.stytch_org_id,
    o.stytch_connection_id,
    o.stytch_connection_name,
    o.created_at,
    o.updated_at
FROM organizations.organizations o
INNER JOIN organizations.accounts a ON o.id = a.organization_id
WHERE a.id = $1;

-- name: CheckAccountPermission :one
SELECT
    a.id,
    a.role,
    a.status,
    o.status as org_status
FROM organizations.accounts a
INNER JOIN organizations.organizations o ON a.organization_id = o.id
WHERE a.id = $1 AND a.organization_id = $2;

-- Statistics queries (useful for admin panels)

-- name: GetOrganizationStats :one
SELECT
    o.id,
    o.slug,
    o.name,
    o.status,
    o.stytch_org_id,
    o.stytch_connection_id,
    o.stytch_connection_name,
    o.created_at,
    o.updated_at,
    COUNT(a.id) as account_count,
    COUNT(CASE WHEN a.status = 'active' THEN 1 END) as active_account_count
FROM organizations.organizations o
LEFT JOIN organizations.accounts a ON o.id = a.organization_id
WHERE o.id = $1
GROUP BY o.id;

-- name: GetAccountStats :one
SELECT
    a.id,
    a.organization_id,
    a.email,
    a.full_name,
    a.stytch_member_id,
    a.stytch_role_id,
    a.stytch_role_slug,
    a.stytch_email_verified,
    a.role,
    a.status,
    a.last_login_at,
    a.created_at,
    a.updated_at,
    o.name as organization_name,
    o.slug as organization_slug
FROM organizations.accounts a
INNER JOIN organizations.organizations o ON a.organization_id = o.id
WHERE a.id = $1;

