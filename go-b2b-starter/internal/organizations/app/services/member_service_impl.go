package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/logger"
)

// rollbackFunc represents a function that can rollback a created resource
type rollbackFunc func(context.Context) error

// rollbackStack manages rollback functions in LIFO order
type rollbackStack []rollbackFunc

// add appends a rollback function to the stack
func (rs *rollbackStack) add(fn rollbackFunc) {
	*rs = append(*rs, fn)
}

// execute runs all rollback functions in reverse order (LIFO)
func (rs rollbackStack) execute(ctx context.Context, logger loggerDomain.Logger) {
	// Execute in reverse order (LIFO - Last In First Out)
	for i := len(rs) - 1; i >= 0; i-- {
		if err := rs[i](ctx); err != nil {
			logger.Error("rollback operation failed", loggerDomain.Fields{
				"step":  i,
				"error": err.Error(),
			})
			// Continue with remaining rollbacks even if one fails
		}
	}
}

type memberService struct {
	authOrgRepo      domain.AuthOrganizationRepository
	authMemberRepo   domain.AuthMemberRepository
	authRoleRepo     domain.AuthRoleRepository
	localOrgRepo     domain.OrganizationRepository
	localAccountRepo domain.AccountRepository
	logger           loggerDomain.Logger
}

func NewMemberService(
	authOrgRepo domain.AuthOrganizationRepository,
	authMemberRepo domain.AuthMemberRepository,
	authRoleRepo domain.AuthRoleRepository,
	localOrgRepo domain.OrganizationRepository,
	localAccountRepo domain.AccountRepository,
	logger loggerDomain.Logger,
) MemberService {
	return &memberService{
		authOrgRepo:      authOrgRepo,
		authMemberRepo:   authMemberRepo,
		authRoleRepo:     authRoleRepo,
		localOrgRepo:     localOrgRepo,
		localAccountRepo: localAccountRepo,
		logger:           logger,
	}
}

// BootstrapOrganizationWithOwner creates a new organization with an initial owner member.
// If any step fails, all previously created resources are automatically rolled back.
func (s *memberService) BootstrapOrganizationWithOwner(
	ctx context.Context,
	req *BootstrapOrganizationRequest,
) (*BootstrapOrganizationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid bootstrap request: %w", err)
	}

	// Always use "admin" role for bootstrap (primary admin user)
	ownerRoleSlug := "admin"

	// Initialize rollback stack for transaction-like behavior
	var rollbacks rollbackStack
	shouldRollback := true

	// Defer rollback execution - runs on function exit if shouldRollback is true
	defer func() {
		if shouldRollback {
			s.logger.Warn("bootstrap failed, executing rollback", loggerDomain.Fields{
				"org_name":       req.OrgDisplayName,
				"rollback_steps": len(rollbacks),
			})
			rollbacks.execute(context.Background(), s.logger)
		}
	}()

	s.logger.Info("starting organization bootstrap", loggerDomain.Fields{
		"org_name":    req.OrgDisplayName,
		"owner_email": req.OwnerEmail,
	})

	// Step 1: Create organization in auth provider
	// Infrastructure layer handles slug generation and duplicate retry logic
	authOrg, err := s.authOrgRepo.CreateOrganization(ctx, &domain.CreateAuthOrganizationRequest{
		DisplayName:         req.OrgDisplayName,
		EmailInvitesAllowed: true,
	})
	if err != nil {
		fmt.Println(err)
		s.logger.Error("failed to create auth organization", loggerDomain.Fields{
			"org_name": req.OrgDisplayName,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Track auth org creation for rollback
	rollbacks.add(func(ctx context.Context) error {
		s.logger.Info("rolling back auth organization", loggerDomain.Fields{
			"auth_org_id": authOrg.OrganizationID,
		})
		return s.authOrgRepo.DeleteOrganization(ctx, authOrg.OrganizationID)
	})

	s.logger.Info("organization created in auth provider", loggerDomain.Fields{
		"auth_org_id": authOrg.OrganizationID,
		"org_slug":    authOrg.Slug,
	})

	// Step 2: Create local organization record.
	s.logger.Info("creating local organization", loggerDomain.Fields{
		"slug":         authOrg.Slug,
		"display_name": authOrg.DisplayName,
	})

	localOrg, err := s.localOrgRepo.Create(ctx, &domain.Organization{
		Slug:   authOrg.Slug,
		Name:   authOrg.DisplayName,
		Status: "active",
	})
	if err != nil {
		s.logger.Error("failed to create local organization", loggerDomain.Fields{
			"org_slug":     authOrg.Slug,
			"display_name": authOrg.DisplayName,
			"status":       "active",
			"error":        err.Error(),
			"error_type":   fmt.Sprintf("%T", err),
		})
		return nil, fmt.Errorf("failed to create local organization: %w", err)
	}

	// Track local org creation for rollback
	rollbacks.add(func(ctx context.Context) error {
		s.logger.Info("rolling back local organization", loggerDomain.Fields{
			"local_org_id": localOrg.ID,
		})
		return s.localOrgRepo.Delete(ctx, localOrg.ID)
	})

	s.logger.Info("local organization created successfully", loggerDomain.Fields{
		"local_org_id": localOrg.ID,
		"slug":         localOrg.Slug,
	})

	if _, err := s.localOrgRepo.UpdateStytchInfo(ctx, localOrg.ID, authOrg.OrganizationID, "", ""); err != nil {
		s.logger.Error("failed to map auth organization locally", loggerDomain.Fields{
			"local_org_id": localOrg.ID,
			"auth_org_id":  authOrg.OrganizationID,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to map auth organization: %w", err)
	}

	// Step 3: Create owner member (no invite).
	createMemberReq := &domain.CreateAuthMemberRequest{
		OrganizationID: authOrg.OrganizationID,
		Email:          req.OwnerEmail,
		Name:           req.OwnerName,
		SendInvite:     false, // No magic link invite
	}

	member, err := s.authMemberRepo.CreateMember(ctx, createMemberReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create owner member: %w", err)
	}

	// Track auth member creation for rollback
	rollbacks.add(func(ctx context.Context) error {
		s.logger.Info("rolling back auth member", loggerDomain.Fields{
			"member_id":   member.MemberID,
			"auth_org_id": authOrg.OrganizationID,
		})
		return s.authMemberRepo.RemoveMembers(ctx, &domain.RemoveAuthMembersRequest{
			OrganizationID: authOrg.OrganizationID,
			MemberIDs:      []string{member.MemberID},
		})
	})

	// Step 4: Assign admin role in auth provider.
	if err := s.authMemberRepo.AssignRoles(ctx, &domain.AssignAuthRolesRequest{
		OrganizationID: authOrg.OrganizationID,
		MemberID:       member.MemberID,
		Roles:          []string{ownerRoleSlug},
	}); err != nil {
		return nil, fmt.Errorf("failed to assign admin role: %w", err)
	}

	role, err := s.authRoleRepo.GetRoleBySlug(ctx, ownerRoleSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch admin role metadata: %w", err)
	}

	// Step 5: Create local account record.
	localAccount, err := s.localAccountRepo.Create(ctx, &domain.Account{
		OrganizationID: localOrg.ID,
		Email:          member.Email,
		FullName:       member.Name,
		Role:           mapRoleSlugToAccountRole(ownerRoleSlug),
		Status:         "active",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create local account: %w", err)
	}

	// Track local account creation for rollback
	rollbacks.add(func(ctx context.Context) error {
		s.logger.Info("rolling back local account", loggerDomain.Fields{
			"account_id":   localAccount.ID,
			"local_org_id": localOrg.ID,
		})
		return s.localAccountRepo.Delete(ctx, localOrg.ID, localAccount.ID)
	})

	if _, err := s.localAccountRepo.UpdateStytchInfo(
		ctx,
		localOrg.ID,
		localAccount.ID,
		member.MemberID,
		role.RoleID,
		ownerRoleSlug,
		member.EmailVerified,
	); err != nil {
		return nil, fmt.Errorf("failed to map auth member locally: %w", err)
	}

	// Success! Disable rollback
	shouldRollback = false

	s.logger.Info("organization bootstrap completed", loggerDomain.Fields{
		"stytch_org_id": authOrg.OrganizationID,
		"owner_member":  member.MemberID,
	})

	return &BootstrapOrganizationResponse{
		OrganizationID: authOrg.OrganizationID,
		OrgSlug:        authOrg.Slug,
		DisplayName:    authOrg.DisplayName,
		OwnerMemberID:  member.MemberID,
		OwnerEmail:     member.Email,
		OwnerName:      member.Name,
		InviteSent:     false, // No invite sent
		MagicLinkSent:  false, // No magic link sent
	}, nil
}

// AddMemberDirect adds a new member to an existing organization without invitation workflows.
func (s *memberService) AddMemberDirect(
	ctx context.Context,
	req *AddMemberRequest,
) (*AddMemberResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid add member request: %w", err)
	}

	roleSlug := strings.ToLower(strings.TrimSpace(req.RoleSlug))
	if roleSlug == "" {
		roleSlug = "member"
	}

	orgID := req.OrgID
	if orgID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}

	localOrgID, err := s.resolveLocalOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if existingAccount, err := s.localAccountRepo.GetByEmail(ctx, localOrgID, req.Email); err == nil {
		s.logger.Warn("member email already exists locally", loggerDomain.Fields{
			"org_id": localOrgID,
			"email":  req.Email,
			"status": existingAccount.Status,
		})
		return nil, domain.ErrAuthMemberAlreadyExists
	} else if !errors.Is(err, domain.ErrAccountNotFound) {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}

	createReq := &domain.CreateAuthMemberRequest{
		OrganizationID: orgID,
		Email:          req.Email,
		Name:           req.Name,
		SendInvite:     false, // No magic link invite
	}

	member, err := s.authMemberRepo.CreateMember(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	if err := s.authMemberRepo.AssignRoles(ctx, &domain.AssignAuthRolesRequest{
		OrganizationID: orgID,
		MemberID:       member.MemberID,
		Roles:          []string{roleSlug},
	}); err != nil {
		return nil, fmt.Errorf("failed to assign member role: %w", err)
	}

	role, err := s.authRoleRepo.GetRoleBySlug(ctx, roleSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role metadata: %w", err)
	}

	localAccount, err := s.localAccountRepo.Create(ctx, &domain.Account{
		OrganizationID: localOrgID,
		Email:          member.Email,
		FullName:       member.Name,
		Role:           mapRoleSlugToAccountRole(roleSlug),
		Status:         "active",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create local account: %w", err)
	}

	if _, err := s.localAccountRepo.UpdateStytchInfo(
		ctx,
		localOrgID,
		localAccount.ID,
		member.MemberID,
		role.RoleID,
		roleSlug,
		member.EmailVerified,
	); err != nil {
		return nil, fmt.Errorf("failed to map auth member locally: %w", err)
	}

	s.logger.Info("member added successfully", loggerDomain.Fields{
		"org_id":      orgID,
		"member_id":   member.MemberID,
		"invite_sent": true,
	})

	return &AddMemberResponse{
		MemberID:   member.MemberID,
		Email:      member.Email,
		Name:       member.Name,
		OrgID:      orgID,
		RoleSlug:   roleSlug,
		InviteSent: true, // Always true (passwordless)
	}, nil
}

// ListOrganizationMembers retrieves all members of an organization.
func (s *memberService) ListOrganizationMembers(
	ctx context.Context,
	orgID string,
) (*ListMembersResponse, error) {
	if orgID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}

	// Retrieve members from repository (no pagination limit)
	members, err := s.authMemberRepo.ListMembers(ctx, orgID, 0, 0)
	if err != nil {
		s.logger.Error("failed to list organization members", loggerDomain.Fields{
			"org_id": orgID,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	// Convert domain members to response info
	memberInfos := make([]*MemberInfo, 0, len(members))
	for _, member := range members {
		memberInfos = append(memberInfos, &MemberInfo{
			MemberID:      member.MemberID,
			Email:         member.Email,
			Name:          member.Name,
			Roles:         member.Roles,
			Status:        member.Status,
			EmailVerified: member.EmailVerified,
			CreatedAt:     member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     member.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	s.logger.Info("members listed successfully", loggerDomain.Fields{
		"org_id": orgID,
		"count":  len(memberInfos),
	})

	return &ListMembersResponse{
		Members: memberInfos,
		Total:   len(memberInfos),
	}, nil
}

// GetCurrentUserProfile retrieves the current authenticated user's profile.
func (s *memberService) GetCurrentUserProfile(
	ctx context.Context,
	orgID, memberID, email string,
) (*ProfileResponse, error) {
	// Validate required parameters
	if orgID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}
	if memberID == "" {
		return nil, domain.ErrAuthMemberIDRequired
	}
	if email == "" {
		return nil, domain.ErrAuthEmailRequired
	}

	// Get member details from auth provider
	member, err := s.authMemberRepo.GetMember(ctx, orgID, memberID)
	if err != nil {
		s.logger.Error("failed to get member details", loggerDomain.Fields{
			"org_id":    orgID,
			"member_id": memberID,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to get member details: %w", err)
	}

	// Get organization details from auth provider
	organization, err := s.authOrgRepo.GetOrganization(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get organization details", loggerDomain.Fields{
			"org_id": orgID,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get organization details: %w", err)
	}

	// Get local organization details (for database ID)
	localOrg, err := s.localOrgRepo.GetByStytchID(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get local organization", loggerDomain.Fields{
			"auth_org_id": orgID,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to get local organization: %w", err)
	}

	// Get local account details (for database ID)
	localAccount, err := s.localAccountRepo.GetByEmail(ctx, localOrg.ID, email)
	if err != nil {
		s.logger.Error("failed to get local account", loggerDomain.Fields{
			"org_id": localOrg.ID,
			"email":  email,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to get local account: %w", err)
	}

	// Build profile response
	profile := &ProfileResponse{
		MemberID:      member.MemberID,
		Email:         member.Email,
		Name:          member.Name,
		Roles:         member.Roles,
		EmailVerified: member.EmailVerified,
		Status:        member.Status,
		Organization: ProfileOrganization{
			OrganizationID: organization.OrganizationID,
			Slug:           organization.Slug,
			Name:           organization.DisplayName,
			Status:         organization.Status,
		},
		AccountID: localAccount.ID,
		CreatedAt: member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: member.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	s.logger.Info("profile retrieved successfully", loggerDomain.Fields{
		"member_id": memberID,
		"org_id":    orgID,
		"email":     email,
	})

	return profile, nil
}

// DeleteOrganizationMember removes a member from the organization
// This deletes from both auth provider and the internal database
// Admin-only operation (permission check done at handler level)
func (s *memberService) DeleteOrganizationMember(
	ctx context.Context,
	orgID, memberID string,
) error {
	if orgID == "" || memberID == "" {
		return fmt.Errorf("organization ID and member ID are required")
	}

	s.logger.Info("deleting organization member", map[string]interface{}{
		"org_id":    orgID,
		"member_id": memberID,
	})

	// Create remove members request
	req := &domain.RemoveAuthMembersRequest{
		OrganizationID: orgID,
		MemberIDs:      []string{memberID},
	}

	// Remove from auth organization
	err := s.authMemberRepo.RemoveMembers(ctx, req)
	if err != nil {
		s.logger.Error("failed to remove member from auth organization", map[string]interface{}{
			"org_id":    orgID,
			"member_id": memberID,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to remove member: %w", err)
	}

	s.logger.Info("member successfully deleted from organization", map[string]interface{}{
		"org_id":    orgID,
		"member_id": memberID,
	})

	return nil
}

// Returns true if email is found in any organization, false otherwise
func (s *memberService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	// Validate email format
	email = strings.TrimSpace(email)
	if email == "" {
		return false, fmt.Errorf("email cannot be empty")
	}

	s.logger.Info("checking email existence", loggerDomain.Fields{
		"email": email,
	})

	// Check using organization repository
	exists, err := s.authOrgRepo.CheckEmailExists(ctx, email)
	if err != nil {
		s.logger.Error("failed to check email existence", loggerDomain.Fields{
			"email": email,
			"error": err.Error(),
		})
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	s.logger.Info("email existence check completed", loggerDomain.Fields{
		"email":  email,
		"exists": exists,
	})

	return exists, nil
}

func (s *memberService) resolveLocalOrganizationID(ctx context.Context, authOrgID string) (int32, error) {
	org, err := s.localOrgRepo.GetByStytchID(ctx, authOrgID)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve local organization: %w", err)
	}
	return org.ID, nil
}

func mapRoleSlugToAccountRole(slug string) string {
	switch strings.ToLower(strings.TrimSpace(slug)) {
	case "owner":
		// Legacy: map owner to admin
		return "admin"
	case "admin":
		return "admin"
	case "approver":
		return "approver"
	case "reviewer":
		// Legacy: map reviewer to approver
		return "approver"
	case "employee", "member":
		return "member"
	default:
		return slug
	}
}
