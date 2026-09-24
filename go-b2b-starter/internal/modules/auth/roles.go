package auth

// Role is a provider-verified assignment. It does not grant permissions by itself.
type Role string

const (
	RoleMember  Role = "member"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"

	// Existing installations may still have these provider role IDs.
	RoleOwner    Role = "owner"
	RoleApprover Role = "approver"
	RoleReviewer Role = "reviewer"
	RoleEmployee Role = "employee"
)

func (r Role) String() string { return string(r) }

// NormalizeRole preserves compatibility with historical and provider-prefixed IDs.
// Permission resolution still uses the original role ID against Stytch's policy.
func NormalizeRole(role string) Role {
	switch Role(role) {
	case "stytch_member", RoleEmployee:
		return RoleMember
	case "stytch_admin", RoleOwner:
		return RoleAdmin
	case "stytch_manager", RoleApprover, RoleReviewer:
		return RoleManager
	default:
		return Role(role)
	}
}
