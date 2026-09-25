package auth

import "strings"

// Permission is a provider-verified resource:action grant, such as org:manage.
type Permission string

func NewPermission(resource, action string) Permission { return Permission(resource + ":" + action) }
func (p Permission) String() string                    { return string(p) }

func (p Permission) Resource() string {
	resource, _, _ := strings.Cut(string(p), ":")
	return resource
}

func (p Permission) Action() string {
	_, action, _ := strings.Cut(string(p), ":")
	return action
}

// MatchesWithWildcard supports resource:*, *:action and *:* grants from Stytch.
func (p Permission) MatchesWithWildcard(other Permission) bool {
	return (p.Resource() == "*" || p.Resource() == other.Resource()) &&
		(p.Action() == "*" || p.Action() == other.Action())
}

func PermissionsToStrings(permissions []Permission) []string {
	result := make([]string, len(permissions))
	for i, p := range permissions {
		result[i] = string(p)
	}
	return result
}
