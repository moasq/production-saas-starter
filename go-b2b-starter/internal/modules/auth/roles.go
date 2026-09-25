package auth

// Role is a live Better Auth membership assignment, never a client-supplied grant.
type Role string

const (
	RoleMember  Role = "member"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
)

func (r Role) String() string { return string(r) }
