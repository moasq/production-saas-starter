-- name: SyncOrganization :one
INSERT INTO organizations.organizations(auth_org_id,polar_customer_external_id,name,slug)
VALUES ($1,$1,$2,$3)
ON CONFLICT (auth_org_id) DO UPDATE SET name=EXCLUDED.name,slug=EXCLUDED.slug
RETURNING *;

-- name: SyncAccount :one
INSERT INTO organizations.accounts(organization_id,email,full_name,role,auth_user_id,auth_member_id)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (organization_id,(lower(email))) DO UPDATE SET full_name=EXCLUDED.full_name,role=EXCLUDED.role,
 auth_user_id=EXCLUDED.auth_user_id,auth_member_id=EXCLUDED.auth_member_id,status='active',last_login_at=CURRENT_TIMESTAMP
RETURNING *;

-- name: GetTenantOrganization :one
SELECT * FROM organizations.organizations WHERE id=$1;

-- name: GetTenantStats :one
SELECT count(*)::bigint AS total_members, count(*) FILTER (WHERE status='active')::bigint AS active_members
FROM organizations.accounts WHERE organization_id=$1;
