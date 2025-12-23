package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/platform/logger/domain"
	stytchcfg "github.com/moasq/go-b2b-starter/internal/platform/stytch"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/rbac"
)

type stytchRoleRepository struct {
	client *stytchcfg.Client
	logger loggerDomain.Logger
}

// NewStytchRoleRepository creates a Stytch-backed role repository.
func NewStytchRoleRepository(client *stytchcfg.Client, logger loggerDomain.Logger) domain.AuthRoleRepository {
	return &stytchRoleRepository{
		client: client,
		logger: logger,
	}
}

func (r *stytchRoleRepository) GetRoleByID(ctx context.Context, roleID string) (*domain.AuthRole, error) {
	if roleID == "" {
		return nil, domain.ErrAuthRoleNotFound
	}

	role, err := r.findRole(ctx, func(role *rbac.PolicyRole) bool {
		return role.RoleID == roleID
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrAuthRoleNotFound
	}
	return role, nil
}

func (r *stytchRoleRepository) GetRoleBySlug(ctx context.Context, slug string) (*domain.AuthRole, error) {
	if slug == "" {
		return nil, domain.ErrAuthRoleNotFound
	}

	role, err := r.findRole(ctx, func(role *rbac.PolicyRole) bool {
		return role.RoleID == slug
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrAuthRoleNotFound
	}
	return role, nil
}

func (r *stytchRoleRepository) ListRoles(ctx context.Context, limit, offset int) ([]*domain.AuthRole, error) {
	policy, err := r.fetchPolicy(ctx)
	if err != nil {
		return nil, err
	}
	if policy == nil || len(policy.Roles) == 0 {
		return nil, nil
	}

	start := offset
	if start < 0 {
		start = 0
	}
	if start > len(policy.Roles) {
		start = len(policy.Roles)
	}

	end := len(policy.Roles)
	if limit > 0 && start+limit < end {
		end = start + limit
	}

	result := make([]*domain.AuthRole, 0, end-start)
	roles := policy.Roles[start:end]
	for i := range roles {
		role := roles[i]
		result = append(result, mapToAuthRole(&role))
	}
	return result, nil
}

func (r *stytchRoleRepository) findRole(ctx context.Context, predicate func(*rbac.PolicyRole) bool) (*domain.AuthRole, error) {
	policy, err := r.fetchPolicy(ctx)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}

	for i := range policy.Roles {
		role := policy.Roles[i]
		if predicate(&role) {
			return mapToAuthRole(&role), nil
		}
	}
	return nil, nil
}

func (r *stytchRoleRepository) fetchPolicy(ctx context.Context) (*rbac.Policy, error) {
	resp, err := r.client.API().RBAC.Policy(ctx, &rbac.PolicyParams{})
	if err != nil {
		return nil, fmt.Errorf("stytch fetch rbac policy: %w", stytchcfg.MapError(err))
	}
	return resp.Policy, nil
}

func mapToAuthRole(src *rbac.PolicyRole) *domain.AuthRole {
	if src == nil {
		return nil
	}

	role := &domain.AuthRole{
		RoleID:      src.RoleID,
		Name:        src.RoleID,
		Description: src.Description,
	}

	if len(src.Permissions) > 0 {
		perms := make([]string, 0, len(src.Permissions))
		for _, perm := range src.Permissions {
			// Actions may be empty for resource-only permissions.
			if len(perm.Actions) == 0 {
				perms = append(perms, perm.ResourceID)
				continue
			}
			for _, action := range perm.Actions {
				perms = append(perms, fmt.Sprintf("%s:%s", perm.ResourceID, action))
			}
		}
		role.Permissions = perms
	}

	return role
}
