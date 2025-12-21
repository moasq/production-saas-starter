package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/logger"
	stytchcfg "github.com/moasq/go-b2b-starter/internal/stytch"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/magiclinks/email"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/organizations"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/organizations/members"
)

type stytchMemberRepository struct {
	client *stytchcfg.Client
	config stytchcfg.Config
	logger loggerDomain.Logger
}

// NewStytchMemberRepository creates a Stytch-backed member repository.
func NewStytchMemberRepository(client *stytchcfg.Client, cfg stytchcfg.Config, logger loggerDomain.Logger) domain.AuthMemberRepository {
	return &stytchMemberRepository{
		client: client,
		config: cfg,
		logger: logger,
	}
}

func (r *stytchMemberRepository) CreateMember(ctx context.Context, req *domain.CreateAuthMemberRequest) (*domain.AuthMember, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create member request: %w", err)
	}

	stytchReq := &members.CreateParams{
		OrganizationID: req.OrganizationID,
		EmailAddress:   req.Email,
		Name:           req.Name,
	}

	resp, err := r.client.API().Organizations.Members.Create(ctx, stytchReq)
	if err != nil {
		r.logger.Error("failed to create member in Stytch", loggerDomain.Fields{
			"org_id": req.OrganizationID,
			"email":  req.Email,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("stytch create member: %w", stytchcfg.MapError(err))
	}

	member := mapToAuthMember(resp.Member)

	return member, nil
}

func (r *stytchMemberRepository) UpdateMember(ctx context.Context, req *domain.UpdateAuthMemberRequest) (*domain.AuthMember, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid update member request: %w", err)
	}

	params := &members.UpdateParams{
		OrganizationID: req.OrganizationID,
		MemberID:       req.MemberID,
	}

	if req.Name != nil {
		params.Name = *req.Name
	}
	if len(req.Roles) > 0 {
		rolesCopy := append([]string(nil), req.Roles...)
		params.Roles = &rolesCopy
	}
	if req.TrustedMeta != nil {
		params.TrustedMetadata = req.TrustedMeta
	}
	if req.UntrustedMeta != nil {
		params.UntrustedMetadata = req.UntrustedMeta
	}

	resp, err := r.client.API().Organizations.Members.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stytch update member: %w", stytchcfg.MapError(err))
	}

	return mapToAuthMember(resp.Member), nil
}

func (r *stytchMemberRepository) GetMember(ctx context.Context, organizationID, memberID string) (*domain.AuthMember, error) {
	if organizationID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}
	if memberID == "" {
		return nil, domain.ErrAuthMemberIDRequired
	}

	resp, err := r.client.API().Organizations.Members.Get(ctx, &members.GetParams{
		OrganizationID: organizationID,
		MemberID:       memberID,
	})
	if err != nil {
		return nil, fmt.Errorf("stytch get member: %w", stytchcfg.MapError(err))
	}

	return mapToAuthMember(resp.Member), nil
}

func (r *stytchMemberRepository) GetMemberByEmail(ctx context.Context, organizationID, emailAddr string) (*domain.AuthMember, error) {
	if organizationID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}
	if emailAddr == "" {
		return nil, domain.ErrAuthEmailRequired
	}

	resp, err := r.client.API().Organizations.Members.Get(ctx, &members.GetParams{
		OrganizationID: organizationID,
		EmailAddress:   emailAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("stytch get member by email: %w", stytchcfg.MapError(err))
	}

	return mapToAuthMember(resp.Member), nil
}

func (r *stytchMemberRepository) ListMembers(ctx context.Context, organizationID string, limit, offset int) ([]*domain.AuthMember, error) {
	if organizationID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}

	params := &members.SearchParams{
		OrganizationIds: []string{organizationID},
	}

	requested := 0
	if limit > 0 {
		requested = limit
	}
	if offset > 0 {
		requested += offset
	}
	if requested > 0 {
		params.Limit = uint32(requested)
	}

	resp, err := r.client.API().Organizations.Members.Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stytch search members: %w", stytchcfg.MapError(err))
	}

	membersList := resp.Members
	if offset > 0 {
		if offset >= len(membersList) {
			membersList = nil
		} else {
			membersList = membersList[offset:]
		}
	}
	if limit > 0 && limit < len(membersList) {
		membersList = membersList[:limit]
	}

	results := make([]*domain.AuthMember, 0, len(membersList))
	for _, m := range membersList {
		results = append(results, mapToAuthMember(m))
	}

	return results, nil
}

func (r *stytchMemberRepository) RemoveMembers(ctx context.Context, req *domain.RemoveAuthMembersRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid remove members request: %w", err)
	}

	for _, memberID := range req.MemberIDs {
		_, err := r.client.API().Organizations.Members.Delete(ctx, &members.DeleteParams{
			OrganizationID: req.OrganizationID,
			MemberID:       memberID,
		})
		if err != nil {
			return fmt.Errorf("stytch delete member %s: %w", memberID, stytchcfg.MapError(err))
		}
	}

	return nil
}

func (r *stytchMemberRepository) AssignRoles(ctx context.Context, req *domain.AssignAuthRolesRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid assign roles request: %w", err)
	}

	updateParams := &members.UpdateParams{
		OrganizationID: req.OrganizationID,
		MemberID:       req.MemberID,
	}
	if len(req.Roles) > 0 {
		rolesCopy := append([]string(nil), req.Roles...)
		updateParams.Roles = &rolesCopy
	}

	_, err := r.client.API().Organizations.Members.Update(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("stytch update member roles: %w", stytchcfg.MapError(err))
	}
	return nil
}

func (r *stytchMemberRepository) SendMagicLink(ctx context.Context, req *domain.SendMagicLinkRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid magic link request: %w", err)
	}

	params := &email.LoginOrSignupParams{
		OrganizationID: req.OrganizationID,
		EmailAddress:   req.Email,
	}

	loginRedirect := strings.TrimSpace(req.LoginRedirectURL)
	if loginRedirect == "" {
		loginRedirect = strings.TrimSpace(r.config.LoginRedirectURL)
	}
	if loginRedirect != "" {
		params.LoginRedirectURL = loginRedirect
	}

	signupRedirect := strings.TrimSpace(req.SignupRedirectURL)
	if signupRedirect == "" {
		signupRedirect = strings.TrimSpace(r.config.InviteRedirectURL)
		if signupRedirect == "" {
			signupRedirect = loginRedirect
		}
	}
	if signupRedirect != "" {
		params.SignupRedirectURL = signupRedirect
	}

	if _, err := r.client.API().MagicLinks.Email.LoginOrSignup(ctx, params); err != nil {
		return fmt.Errorf("stytch send magic link: %w", stytchcfg.MapError(err))
	}

	return nil
}

func mapToAuthMember(src organizations.Member) *domain.AuthMember {
	var createdAt, updatedAt time.Time
	if src.CreatedAt != nil {
		createdAt = src.CreatedAt.UTC()
	}
	if src.UpdatedAt != nil {
		updatedAt = src.UpdatedAt.UTC()
	}

	roleIDs := make([]string, 0, len(src.Roles))
	for _, role := range src.Roles {
		if role.RoleID != "" {
			roleIDs = append(roleIDs, role.RoleID)
		}
	}

	return &domain.AuthMember{
		MemberID:       src.MemberID,
		OrganizationID: src.OrganizationID,
		Email:          src.EmailAddress,
		Name:           src.Name,
		Roles:          roleIDs,
		Status:         src.Status,
		EmailVerified:  src.EmailAddressVerified,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}
