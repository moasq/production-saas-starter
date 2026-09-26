package auth

// These explicit application grants come from the repository-owned policy in
// next_b2b_starter/lib/auth/rbac.ts. Go enforces the current bridge result and
// never restores a removed permission from a membership role name.
const (
	PermOrgView   Permission = "org:view"
	PermOrgManage Permission = "org:manage"
)
