package stytch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/internal/logger"
	"github.com/moasq/go-b2b-starter/internal/redis"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/rbac"
)

const (
	// Redis cache key for RBAC policy
	rbacPolicyCacheKey = "stytch:rbac:policy"
	// Cache TTL matches Stytch SDK default (5 minutes)
	rbacPolicyCacheTTL = 5 * time.Minute
)

// RBACPolicyService fetches and caches Stytch RBAC policy
type RBACPolicyService struct {
	client *Client
	redis  redis.Client
	logger logger.Logger
}

func NewRBACPolicyService(
	client *Client,
	redisClient redis.Client,
	logger logger.Logger,
) *RBACPolicyService {
	return &RBACPolicyService{
		client: client,
		redis:  redisClient,
		logger: logger,
	}
}

// GetRolePermissions returns all permissions for a given role from Stytch RBAC policy
// Returns permissions in "resource:action" format (e.g., "invoice:create")
func (s *RBACPolicyService) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	// Normalize role ID (remove stytch_ prefix)
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
	return nil, nil
}

// getPolicy fetches policy from Redis cache or Stytch API
func (s *RBACPolicyService) getPolicy(ctx context.Context) (*rbac.Policy, error) {
	// Try cache first
	cached, err := s.redis.Get(ctx, rbacPolicyCacheKey)
	if err == nil && cached != "" {
		var policy rbac.Policy
		if err := json.Unmarshal([]byte(cached), &policy); err == nil {
			s.logger.Debug("RBAC policy fetched from cache", logger.Fields{})
			return &policy, nil
		} else {
			s.logger.Warn("Failed to unmarshal cached RBAC policy, fetching from Stytch", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	// Cache miss or error - fetch from Stytch
	policy, err := s.fetchPolicyFromStytch(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the policy
	s.cachePolicy(ctx, policy)

	return policy, nil
}

// fetchPolicyFromStytch fetches RBAC policy from Stytch API
func (s *RBACPolicyService) fetchPolicyFromStytch(ctx context.Context) (*rbac.Policy, error) {
	s.logger.Info("Fetching RBAC policy from Stytch", logger.Fields{})

	resp, err := s.client.API().RBAC.Policy(ctx, &rbac.PolicyParams{})
	if err != nil {
		return nil, fmt.Errorf("stytch RBAC policy API call failed: %w", err)
	}

	if resp.Policy == nil {
		return nil, fmt.Errorf("stytch returned empty policy")
	}

	s.logger.Info("Successfully fetched RBAC policy from Stytch", logger.Fields{
		"roles_count":     len(resp.Policy.Roles),
		"resources_count": len(resp.Policy.Resources),
	})

	return resp.Policy, nil
}

// cachePolicy stores policy in Redis
func (s *RBACPolicyService) cachePolicy(ctx context.Context, policy *rbac.Policy) {
	data, err := json.Marshal(policy)
	if err != nil {
		s.logger.Warn("Failed to marshal RBAC policy for caching", logger.Fields{
			"error": err.Error(),
		})
		return
	}

	if err := s.redis.Set(ctx, rbacPolicyCacheKey, string(data), rbacPolicyCacheTTL); err != nil {
		s.logger.Warn("Failed to cache RBAC policy in Redis", logger.Fields{
			"error": err.Error(),
		})
		// Non-fatal error, continue without cache
	} else {
		s.logger.Debug("RBAC policy cached in Redis", logger.Fields{
			"ttl": rbacPolicyCacheTTL.String(),
		})
	}
}

// convertPermissions converts Stytch permission format to flat list, expanding wildcards
//
//	Input: []PolicyRolePermission{
//	  {ResourceID: "Invoice", Actions: ["view", "create"]},
//	  {ResourceID: "approval", Actions: ["*"]},
//	}, policy with Resources
//
// Output: ["invoice:view", "invoice:create", "approval:view", "approval:approve", "approval:assign"]
// (wildcards expanded to all actions defined in policy for that resource)
func (s *RBACPolicyService) convertPermissions(permissions []rbac.PolicyRolePermission, policy *rbac.Policy) []string {
	if len(permissions) == 0 {
		return nil
	}

	result := make([]string, 0, len(permissions)*5) // Estimate 5 actions per resource

	for _, perm := range permissions {
		resourceID := strings.ToLower(perm.ResourceID)

		// Handle empty or invalid resource
		if resourceID == "" {
			continue
		}

		// Expand wildcard actions using resource definitions from policy
		expandedActions := s.expandWildcardActions(perm.ResourceID, perm.Actions, policy)

		// Convert each action to "resource:action" format
		for _, action := range expandedActions {
			if action == "" {
				continue
			}
			actionLower := strings.ToLower(action)
			result = append(result, fmt.Sprintf("%s:%s", resourceID, actionLower))
		}
	}

	return result
}

// expandWildcardActions expands wildcard (*) to all resource actions from Stytch policy
// This ensures permissions come entirely from Stytch configuration, not local code
func (s *RBACPolicyService) expandWildcardActions(resourceID string, actions []string, policy *rbac.Policy) []string {
	// Check if actions contain wildcard
	hasWildcard := false
	for _, action := range actions {
		if action == "*" {
			hasWildcard = true
			break
		}
	}

	// If no wildcard, return actions as-is
	if !hasWildcard {
		return actions
	}

	// Find resource definition in policy to get all possible actions
	for _, resource := range policy.Resources {
		if strings.EqualFold(resource.ResourceID, resourceID) {
			// Return all actions defined for this resource in Stytch
			if len(resource.Actions) > 0 {
				s.logger.Debug("Expanded wildcard permission", logger.Fields{
					"resource":      resourceID,
					"actions_count": len(resource.Actions),
					"actions":       resource.Actions,
				})
				return resource.Actions
			}
		}
	}

	// Resource not found in policy, keep wildcard as-is (shouldn't happen)
	s.logger.Warn("Resource not found in policy, keeping wildcard", logger.Fields{
		"resource": resourceID,
	})
	return actions
}

// normalizeRoleID removes common prefixes from role IDs
// "stytch_member" -> "stytch_member"
// "owner" -> "owner"
// "stytch_admin" -> "stytch_admin"
func normalizeRoleID(roleID string) string {
	roleID = strings.TrimSpace(roleID)
	// Keep stytch_ prefix for default roles, but remove "member" suffix
	roleID = strings.TrimPrefix(roleID, "member")
	roleID = strings.TrimSpace(roleID)
	return roleID
}
