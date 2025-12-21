package services

import (
	"context"
	"fmt"
	"strings"
)

// MemberService defines the core authentication and member management operations
// This interface focuses on organization bootstrap and member operations
type MemberService interface {
	// BootstrapOrganizationWithOwner creates a new organization with an initial owner user
	// This is the primary signup flow for new organizations
	BootstrapOrganizationWithOwner(ctx context.Context, req *BootstrapOrganizationRequest) (*BootstrapOrganizationResponse, error)

	// AddMemberDirect adds a new member to an existing organization without invitation
	// Creates the user if they don't exist, then adds them to the organization with specified roles
	AddMemberDirect(ctx context.Context, req *AddMemberRequest) (*AddMemberResponse, error)

	// ListOrganizationMembers retrieves all members of an organization
	// Returns a list of members with their details including roles and status
	ListOrganizationMembers(ctx context.Context, orgID string) (*ListMembersResponse, error)

	// GetCurrentUserProfile retrieves the current authenticated user's profile
	// Returns comprehensive profile information including member, organization, and account details
	GetCurrentUserProfile(ctx context.Context, orgID, memberID, email string) (*ProfileResponse, error)

	// DeleteOrganizationMember removes a member from the organization (admin only)
	// Deletes from both auth provider and internal database
	DeleteOrganizationMember(ctx context.Context, orgID, memberID string) error

	// CheckEmailExists checks if an email exists in the system
	// Returns true if email is found, false otherwise
	// Used for login flow to verify if user has an account
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

// BootstrapOrganizationRequest represents the request to create a new organization with an owner
type BootstrapOrganizationRequest struct {
	// Organization details
	OrgDisplayName string `json:"org_display_name" binding:"required"`

	// Owner member details
	OwnerEmail string `json:"owner_email" binding:"required,email"`
	OwnerName  string `json:"owner_name" binding:"required"`
}

// Validate performs business validation on the bootstrap request
func (r *BootstrapOrganizationRequest) Validate() error {
	if strings.TrimSpace(r.OrgDisplayName) == "" {
		return fmt.Errorf("organization display name cannot be empty")
	}
	if strings.TrimSpace(r.OwnerEmail) == "" {
		return fmt.Errorf("owner email cannot be empty")
	}
	if strings.TrimSpace(r.OwnerName) == "" {
		return fmt.Errorf("owner name cannot be empty")
	}
	return nil
}

// BootstrapOrganizationResponse represents the response after organization bootstrap
type BootstrapOrganizationResponse struct {
	OrganizationID string `json:"organization_id"`
	OrgSlug        string `json:"org_slug"`
	DisplayName    string `json:"display_name"`
	OwnerMemberID  string `json:"owner_member_id"`
	OwnerEmail     string `json:"owner_email"`
	OwnerName      string `json:"owner_name"`
	InviteSent     bool   `json:"invite_sent"`
	MagicLinkSent  bool   `json:"magic_link_sent"`
}

// AddMemberRequest represents the request to add a member to an organization
type AddMemberRequest struct {
	// Organization context (populated by handler from JWT middleware, not from request body)
	OrgID string `json:"-"`

	// Member user details
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`

	// Role assignment (single role per member)
	RoleSlug string `json:"role_slug"`
}

// Validate performs business validation on the add member request
func (r *AddMemberRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	// Note: OrgID is validated by handler (extracted from JWT middleware)
	return nil
}

// AddMemberResponse represents the response after adding a member
type AddMemberResponse struct {
	MemberID   string `json:"member_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	OrgID      string `json:"org_id"`
	RoleSlug   string `json:"role_slug"`
	InviteSent bool   `json:"invite_sent"`
}

// MemberInfo represents a member in the list response
type MemberInfo struct {
	MemberID      string   `json:"member_id"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles"`
	Status        string   `json:"status"`
	EmailVerified bool     `json:"email_verified"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ListMembersResponse represents the response for listing organization members
type ListMembersResponse struct {
	Members []*MemberInfo `json:"members"`
	Total   int           `json:"total"`
}

// ProfileResponse represents the current user's profile information
// This is a composite response combining auth provider member data + internal account + organization
type ProfileResponse struct {
	// Auth provider member details
	MemberID      string   `json:"member_id"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles"`
	Permissions   []string `json:"permissions"`
	EmailVerified bool     `json:"email_verified"`
	Status        string   `json:"status"`

	// Organization details
	Organization ProfileOrganization `json:"organization"`

	// Internal account details
	AccountID int32  `json:"account_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ProfileOrganization represents organization info in profile response
type ProfileOrganization struct {
	OrganizationID string `json:"organization_id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	Status         string `json:"status"`
}

// CheckEmailRequest represents the request to check if an email exists
// Used for login flow to verify if user has an account
type CheckEmailRequest struct {
	Email string `form:"email" binding:"required,email"`
}
