package services

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/app/organizations/domain"
)

type organizationService struct {
	orgRepo     domain.OrganizationRepository
	accountRepo domain.AccountRepository
}

func NewOrganizationService(orgRepo domain.OrganizationRepository, accountRepo domain.AccountRepository) OrganizationService {
	return &organizationService{
		orgRepo:     orgRepo,
		accountRepo: accountRepo,
	}
}

func (s *organizationService) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*domain.Organization, error) {
	// Create organization
	org := &domain.Organization{
		Slug:                 req.Slug,
		Name:                 req.Name,
		Status:               "active",
		StytchOrgID:          req.StytchOrgID,
		StytchConnectionID:   req.StytchConnectionID,
		StytchConnectionName: req.StytchConnectionName,
	}

	createdOrg, err := s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	if req.StytchOrgID != "" || req.StytchConnectionID != "" || req.StytchConnectionName != "" {
		createdOrg.StytchOrgID = req.StytchOrgID
		createdOrg.StytchConnectionID = req.StytchConnectionID
		createdOrg.StytchConnectionName = req.StytchConnectionName
		createdOrg, err = s.orgRepo.Update(ctx, createdOrg)
		if err != nil {
			return nil, fmt.Errorf("failed to persist organization Stytch metadata: %w", err)
		}
	}

	// Create admin account (primary admin user)
	adminAccount := &domain.Account{
		OrganizationID: createdOrg.ID,
		Email:          req.OwnerEmail,
		FullName:       req.OwnerName,
		Role:           "admin",
		Status:         "active",
	}

	_, err = s.accountRepo.Create(ctx, adminAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin account: %w", err)
	}

	return createdOrg, nil
}

func (s *organizationService) GetOrganization(ctx context.Context, orgID int32) (*domain.Organization, error) {
	return s.orgRepo.GetByID(ctx, orgID)
}

func (s *organizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	return s.orgRepo.GetBySlug(ctx, slug)
}

func (s *organizationService) GetOrganizationByStytchID(ctx context.Context, stytchOrgID string) (*domain.Organization, error) {
	return s.orgRepo.GetByStytchID(ctx, stytchOrgID)
}

func (s *organizationService) GetOrganizationByUserEmail(ctx context.Context, email string) (*domain.Organization, error) {
	return s.orgRepo.GetByUserEmail(ctx, email)
}

func (s *organizationService) UpdateOrganization(ctx context.Context, orgID int32, req *UpdateOrganizationRequest) (*domain.Organization, error) {
	// Get existing organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Update fields
	org.Name = req.Name
	org.Status = req.Status
	if req.StytchOrgID != "" {
		org.StytchOrgID = req.StytchOrgID
	}
	if req.StytchConnectionID != "" {
		org.StytchConnectionID = req.StytchConnectionID
	}
	if req.StytchConnectionName != "" {
		org.StytchConnectionName = req.StytchConnectionName
	}

	return s.orgRepo.Update(ctx, org)
}

func (s *organizationService) ListOrganizations(ctx context.Context, req *ListOrganizationsRequest) (*ListOrganizationsResponse, error) {
	organizations, err := s.orgRepo.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// For simplicity, we're not implementing total count yet
	// In production, you'd want a separate query for total count
	total := int32(len(organizations))

	return &ListOrganizationsResponse{
		Organizations: organizations,
		Total:         total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}, nil
}

func (s *organizationService) GetOrganizationStats(ctx context.Context, orgID int32) (*domain.OrganizationStats, error) {
	return s.orgRepo.GetStats(ctx, orgID)
}

func (s *organizationService) CreateAccount(ctx context.Context, orgID int32, req *CreateAccountRequest) (*domain.Account, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	account := &domain.Account{
		OrganizationID:      orgID,
		Email:               req.Email,
		FullName:            req.FullName,
		StytchMemberID:      req.StytchMemberID,
		StytchRoleID:        req.StytchRoleID,
		StytchRoleSlug:      req.StytchRoleSlug,
		StytchEmailVerified: req.StytchEmailVerified,
		Role:                req.Role,
		Status:              "active",
	}

	return s.accountRepo.Create(ctx, account)
}

func (s *organizationService) GetAccount(ctx context.Context, orgID, accountID int32) (*domain.Account, error) {
	return s.accountRepo.GetByID(ctx, orgID, accountID)
}

func (s *organizationService) GetAccountByEmail(ctx context.Context, orgID int32, email string) (*domain.Account, error) {
	return s.accountRepo.GetByEmail(ctx, orgID, email)
}

func (s *organizationService) ListAccounts(ctx context.Context, orgID int32) ([]*domain.Account, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return s.accountRepo.ListByOrganization(ctx, orgID)
}

func (s *organizationService) UpdateAccount(ctx context.Context, orgID, accountID int32, req *UpdateAccountRequest) (*domain.Account, error) {
	// Get existing account
	account, err := s.accountRepo.GetByID(ctx, orgID, accountID)
	if err != nil {
		return nil, err
	}

	// Update fields
	account.FullName = req.FullName
	account.Role = req.Role
	account.Status = req.Status
	if req.StytchRoleID != "" {
		account.StytchRoleID = req.StytchRoleID
	}
	if req.StytchRoleSlug != "" {
		account.StytchRoleSlug = req.StytchRoleSlug
	}
	if req.StytchEmailVerified != nil {
		account.StytchEmailVerified = *req.StytchEmailVerified
	}

	return s.accountRepo.Update(ctx, account)
}

func (s *organizationService) DeleteAccount(ctx context.Context, orgID, accountID int32) error {
	return s.accountRepo.Delete(ctx, orgID, accountID)
}

func (s *organizationService) UpdateAccountLastLogin(ctx context.Context, orgID, accountID int32) (*domain.Account, error) {
	return s.accountRepo.UpdateLastLogin(ctx, orgID, accountID)
}

func (s *organizationService) CheckAccountPermission(ctx context.Context, orgID, accountID int32) (*domain.AccountPermission, error) {
	return s.accountRepo.CheckPermission(ctx, orgID, accountID)
}

func (s *organizationService) GetAccountStats(ctx context.Context, accountID int32) (*domain.AccountStats, error) {
	return s.accountRepo.GetStats(ctx, accountID)
}
