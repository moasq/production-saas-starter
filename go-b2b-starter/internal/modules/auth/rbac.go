package auth

// The starter requires these provider permissions. Stytch RBAC policy remains
// authoritative; the application never grants permissions from a local role map.
const (
	PermOrgView   Permission = "org:view"
	PermOrgManage Permission = "org:manage"
)
