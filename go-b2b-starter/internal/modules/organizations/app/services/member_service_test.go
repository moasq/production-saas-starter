package services

import (
	"context"
	"errors"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"testing"
)

type orgFake struct{ domain.OrganizationRepository }

func (orgFake) GetByStytchID(context.Context, string) (*domain.Organization, error) {
	return &domain.Organization{ID: 7}, nil
}

type accountsFake struct {
	domain.AccountRepository
	account    *domain.Account
	failDelete bool
}

func (a *accountsFake) ListByOrganization(_ context.Context, orgID int32) ([]*domain.Account, error) {
	if orgID != 7 {
		panic("wrong tenant")
	}
	if a.account == nil {
		return nil, nil
	}
	return []*domain.Account{a.account}, nil
}
func (a *accountsFake) Delete(_ context.Context, orgID, id int32) error {
	if orgID != 7 {
		panic("wrong tenant")
	}
	if a.failDelete {
		a.failDelete = false
		return errors.New("temporary database error")
	}
	a.account = nil
	return nil
}
func (a *accountsFake) GetByEmail(context.Context, int32, string) (*domain.Account, error) {
	if a.account == nil {
		return nil, domain.ErrAccountNotFound
	}
	return a.account, nil
}
func (a *accountsFake) Create(_ context.Context, account *domain.Account) (*domain.Account, error) {
	account.ID = 10
	a.account = account
	return account, nil
}
func (a *accountsFake) UpdateStytchInfo(_ context.Context, orgID, id int32, memberID, roleID, roleSlug string, verified bool) (*domain.Account, error) {
	a.account.StytchMemberID = memberID
	return a.account, nil
}

type providerFake struct {
	domain.AuthMemberRepository
	deleted, created, sent int
	failEmail              bool
	failRole               bool
}

func (p *providerFake) RemoveMembers(context.Context, *domain.RemoveAuthMembersRequest) error {
	p.deleted++
	return nil
}
func (p *providerFake) CreateMember(_ context.Context, req *domain.CreateAuthMemberRequest) (*domain.AuthMember, error) {
	p.created++
	return &domain.AuthMember{MemberID: "new-member", OrganizationID: req.OrganizationID, Email: req.Email, Name: req.Name}, nil
}
func (p *providerFake) AssignRoles(context.Context, *domain.AssignAuthRolesRequest) error {
	if p.failRole {
		return errors.New("role configuration missing")
	}
	return nil
}
func (p *providerFake) SendMagicLink(context.Context, *domain.SendMagicLinkRequest) error {
	p.sent++
	if p.failEmail {
		return errors.New("email unavailable")
	}
	return nil
}

type roleFake struct{ domain.AuthRoleRepository }

func (roleFake) GetRoleBySlug(context.Context, string) (*domain.AuthRole, error) {
	return &domain.AuthRole{RoleID: "member"}, nil
}

func TestMemberDeleteCanRetryAndReinvite(t *testing.T) {
	accounts := &accountsFake{account: &domain.Account{ID: 3, OrganizationID: 7, Email: "member@example.com", StytchMemberID: "old-member"}, failDelete: true}
	provider := &providerFake{}
	service := &memberService{localOrgRepo: orgFake{}, localAccountRepo: accounts, authMemberRepo: provider, authRoleRepo: roleFake{}, logger: logger.New()}
	if err := service.DeleteOrganizationMember(context.Background(), "org-test", "old-member"); err == nil {
		t.Fatal("local failure reported success")
	}
	if err := service.DeleteOrganizationMember(context.Background(), "org-test", "old-member"); err != nil {
		t.Fatal(err)
	}
	if accounts.account != nil {
		t.Fatal("deleted member left local account")
	}
	result, err := service.AddMemberDirect(context.Background(), &AddMemberRequest{OrgID: "org-test", Email: "member@example.com", Name: "Member", RoleSlug: "member"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.InviteSent || provider.created != 1 || provider.sent != 1 || accounts.account.StytchMemberID != "new-member" {
		t.Fatal("reinvite did not recreate member and deliver invitation")
	}
}

func TestInvitationFailureDoesNotClaimEmailWasSent(t *testing.T) {
	provider := &providerFake{failEmail: true}
	accounts := &accountsFake{}
	service := &memberService{localOrgRepo: orgFake{}, localAccountRepo: accounts, authMemberRepo: provider, authRoleRepo: roleFake{}, logger: logger.New()}
	result, err := service.AddMemberDirect(context.Background(), &AddMemberRequest{OrgID: "org-test", Email: "member@example.com", Name: "Member", RoleSlug: "member"})
	if err != nil {
		t.Fatal(err)
	}
	if result.InviteSent || accounts.account == nil {
		t.Fatal("invitation failure must preserve created member and report delivery failure")
	}
}

func TestFailedRoleAssignmentRollsBackProviderMember(t *testing.T) {
	provider := &providerFake{failRole: true}
	accounts := &accountsFake{}
	service := &memberService{localOrgRepo: orgFake{}, localAccountRepo: accounts, authMemberRepo: provider, authRoleRepo: roleFake{}, logger: logger.New()}
	if _, err := service.AddMemberDirect(context.Background(), &AddMemberRequest{OrgID: "org-test", Email: "member@example.com", Name: "Member", RoleSlug: "member"}); err == nil {
		t.Fatal("role failure accepted")
	}
	if provider.deleted != 1 || accounts.account != nil || provider.sent != 0 {
		t.Fatal("failed invite left partially created member")
	}
}
