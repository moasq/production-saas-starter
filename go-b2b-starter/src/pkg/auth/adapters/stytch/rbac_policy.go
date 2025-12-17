package stytch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/pkg/auth"
	"github.com/moasq/go-b2b-starter/pkg/logger"
	"github.com/moasq/go-b2b-starter/pkg/redis"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/b2bstytchapi"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/rbac"
)

const (
	// Redis cache key for RBAC policy
	rbacPolicyCacheKey = "auth:stytch:rbac:policy"
	// Cache TTL matches Stytch SDK default (5 minutes)
	rbacPolicyCacheTTL = 5 * time.Minute
)

// RBACPolicyService fetches and caches the Stytch RBAC policy.
//
// It retrieves the role-permission mappings from Stytch and caches them
// in Redis to avoid API calls on every request.
type RBACPolicyService struct {
	client *b2bstytchapi.API
	redis  redis.Client
	logger logger.Logger
}

func NewRBACPolicyService(client *b2bstytchapi.API, redisClient redis.Client, logger logger.Logger) *RBACPolicyService {
	return &RBACPolicyService{
		client: client,
		redis:  redisClient,
		logger: logger,
	}
}

// GetRolePermissions returns all permissions for a given role from Stytch RBAC policy.
//
// Returns permissions in "resource:action" format (e.g., "invoice:create").
func (s *RBACPolicyService) GetRolePermissions(ctx context.Context, roleID string) ([]auth.Permission, error) {
	// Normalize role ID
	normalizedRoleID := normalizeRoleID(roleID)

	// Get policy from cache or Stytch
	policy, err := s.getPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get RBAC policy: %w", err)
	}

	// Find role in policy
	for _, role := range policy.Roles {
		if strings.EqualFold(role.RoleID, normalizedRoleID) {
			return s.convertPermissions(role.Permissions, policy), nil
		}
	}

	// Role not found in policy
	s.logger.Debug("role not found in Stytch RBAC policy", logger.Fields{
		"role_id":    roleID,
		"normalized": normalizedRoleID,
	})
	return nil, nil
}

// getPolicy fetches policy from Redis cache or Stytch API.
func (s *RBACPolicyService) getPolicy(ctx context.Context) (*rbac.Policy, error) {
	// Try cache first
	cached, err := s.redis.Get(ctx, rbacPolicyCacheKey)
	if err == nil && cached != "" {
		var policy rbac.Policy
		if unmarshalErr := json.Unmarshal([]byte(cached), &policy); unmarshalErr == nil {
			s.logger.Debug("RBAC policy fetched from cache", logger.Fields{})
			return &policy, nil
		} else {
			s.logger.Warn("failed to unmarshal cached RBAC policy", logger.Fields{
				"error": unmarshalErr.Error(),
			})
		}
	}

	// Cache miss - fetch from Stytch
	policy, err := s.fetchPolicyFromStytch(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the policy
	s.cachePolicy(ctx, policy)

	return policy, nil
}

// fetchPolicyFromStytch fetches RBAC policy from Stytch API.
func (s *RBACPolicyService) fetchPolicyFromStytch(ctx context.Context) (*rbac.Policy, error) {
	s.logger.Info("fetching RBAC policy from Stytch", logger.Fields{})

	resp, err := s.client.RBAC.Policy(ctx, &rbac.PolicyParams{})
	if err != nil {
		return nil, fmt.Errorf("stytch RBAC policy API call failed: %w", err)
	}

	if resp.Policy == nil {
		return nil, fmt.Errorf("stytch returned empty policy")
	}

	s.logger.Info("successfully fetched RBAC policy", logger.Fields{
		"roles_count":     len(resp.Policy.Roles),
		"resources_count": len(resp.Policy.Resources),
	})

	return resp.Policy, nil
}

// cachePolicy stores policy in Redis.
func (s *RBACPolicyService) cachePolicy(ctx context.Context, policy *rbac.Policy) {
	data, err := json.Marshal(policy)
	if err != nil {
		s.logger.Warn("failed to marshal RBAC policy for caching", logger.Fields{
			"error": err.Error(),
		})
		return
	}

	if err := s.redis.Set(ctx, rbacPolicyCacheKey, string(data), rbacPolicyCacheTTL); err != nil {
		s.logger.Warn("failed to cache RBAC policy in Redis", logger.Fields{
			"error": err.Error(),
		})
	}
}

// convertPermissions converts Stytch permission format to auth.Permission slice.
//
// Handles wildcard expansion:
//
//	Input: []PolicyRolePermission{{ResourceID: "invoice", Actions: ["view", "create"]}}
//	Output: [Permission("invoice:view"), Permission("invoice:create")]
func (s *RBACPolicyService) convertPermissions(permissions []rbac.PolicyRolePermission, policy *rbac.Policy) []auth.Permission {
	if len(permissions) == 0 {
		return nil
	}

	result := make([]auth.Permission, 0, len(permissions)*5)

	for _, perm := range permissions {
		resourceID := strings.ToLower(perm.ResourceID)
		if resourceID == "" {
			continue
		}

		// Expand wildcard actions
		expandedActions := s.expandWildcardActions(perm.ResourceID, perm.Actions, policy)

		// Convert each action to Permission
		for _, action := range expandedActions {
			if action == "" {
				continue
			}
			result = append(result, auth.NewPermission(resourceID, strings.ToLower(action)))
		}
	}

	return result
}

// expandWildcardActions expands wildcard (*) to all resource actions from policy.
func (s *RBACPolicyService) expandWildcardActions(resourceID string, actions []string, policy *rbac.Policy) []string {
	// Check if actions contain wildcard
	hasWildcard := false
	for _, action := range actions {
		if action == "*" {
			hasWildcard = true
			break
		}
	}

	if !hasWildcard {
		return actions
	}

	// Find resource definition to get all actions
	for _, resource := range policy.Resources {
		if strings.EqualFold(resource.ResourceID, resourceID) {
			if len(resource.Actions) > 0 {
				s.logger.Debug("expanded wildcard permission", logger.Fields{
					"resource":      resourceID,
					"actions_count": len(resource.Actions),
				})
				return resource.Actions
			}
		}
	}

	// Resource not found, keep wildcard as-is
	return actions
}

// normalizeRoleID removes common prefixes from role IDs.
func normalizeRoleID(roleID string) string {
	roleID = strings.TrimSpace(roleID)
	roleID = strings.TrimPrefix(roleID, "stytch_")
	return roleID
}
