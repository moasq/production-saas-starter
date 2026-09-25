package services

import (
	"context"
	"encoding/json"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"
	"testing"
)

type organizationUpdateFake struct {
	domain.OrganizationRepository
	org *domain.Organization
}

func (f *organizationUpdateFake) GetByID(context.Context, int32) (*domain.Organization, error) {
	return f.org, nil
}
func (f *organizationUpdateFake) Update(_ context.Context, org *domain.Organization) (*domain.Organization, error) {
	f.org = org
	return org, nil
}
func TestOrganizationUpdateCannotReplaceProviderLink(t *testing.T) {
	repo := &organizationUpdateFake{org: &domain.Organization{ID: 7, Name: "Before", Status: "active", StytchOrgID: "verified-org", StytchConnectionID: "verified-connection"}}
	var req UpdateOrganizationRequest
	if err := json.Unmarshal([]byte(`{"name":"After","status":"active","stytch_org_id":"attacker-org","stytch_connection_id":"attacker-connection"}`), &req); err != nil {
		t.Fatal(err)
	}
	updated, err := NewOrganizationService(repo, nil).UpdateOrganization(context.Background(), 7, &req)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "After" || updated.StytchOrgID != "verified-org" || updated.StytchConnectionID != "verified-connection" {
		t.Fatal("public update altered provider linkage")
	}
}
