package auth

import (
	"context"
	"fmt"
)

// OrganizationLookup is the minimal interface needed to resolve organizations.
//
// This interface should be implemented by your organization repository.
// It abstracts the specific repository implementation from the auth package.
type OrganizationLookup interface {
	// GetByStytchID returns an organization by Stytch organization ID.
	// The returned value must have an ID field (int32).
	GetByStytchID(ctx context.Context, stytchOrgID string) (OrganizationEntity, error)
}

// OrganizationEntity is the minimal interface for an organization entity.
type OrganizationEntity interface {
	GetID() int32
}

// AccountLookup is the minimal interface needed to resolve accounts.
//
// This interface should be implemented by your account repository.
// It abstracts the specific repository implementation from the auth package.
type AccountLookup interface {
	// GetByEmail returns an account by email within an organization.
	// The returned value must have an ID field (int32).
	GetByEmail(ctx context.Context, orgID int32, email string) (AccountEntity, error)
}

// AccountEntity is the minimal interface for an account entity.
type AccountEntity interface {
	GetID() int32
}

// NewOrganizationResolver creates an OrganizationResolver from an OrganizationLookup.
//
// This is a convenience function for creating resolvers from repositories
// that implement the OrganizationLookup interface.
//
// # Usage
//
//	// In your organizations module provider:
//	container.Provide(func(repo domain.OrganizationRepository) auth.OrganizationResolver {
//	    return auth.NewOrganizationResolver(repo)
//	})
func NewOrganizationResolver(lookup OrganizationLookup) OrganizationResolver {
	return &orgResolverAdapter{lookup: lookup}
}

// NewAccountResolver creates an AccountResolver from an AccountLookup.
//
// # Usage
//
//	// In your organizations module provider:
//	container.Provide(func(repo domain.AccountRepository) auth.AccountResolver {
//	    return auth.NewAccountResolver(repo)
//	})
func NewAccountResolver(lookup AccountLookup) AccountResolver {
	return &accResolverAdapter{lookup: lookup}
}

// orgResolverAdapter adapts OrganizationLookup to OrganizationResolver.
type orgResolverAdapter struct {
	lookup OrganizationLookup
}

func (a *orgResolverAdapter) ResolveByProviderID(ctx context.Context, providerOrgID string) (int32, error) {
	org, err := a.lookup.GetByStytchID(ctx, providerOrgID)
	if err != nil {
		return 0, fmt.Errorf("organization not found for provider ID %s: %w", providerOrgID, err)
	}
	return org.GetID(), nil
}

// accResolverAdapter adapts AccountLookup to AccountResolver.
type accResolverAdapter struct {
	lookup AccountLookup
}

func (a *accResolverAdapter) ResolveByEmail(ctx context.Context, orgID int32, email string) (int32, error) {
	acc, err := a.lookup.GetByEmail(ctx, orgID, email)
	if err != nil {
		return 0, fmt.Errorf("account not found for email %s in org %d: %w", email, orgID, err)
	}
	return acc.GetID(), nil
}

// SimpleOrganization is a simple implementation of OrganizationEntity.
// Use this if your domain entity doesn't already implement GetID().
type SimpleOrganization struct {
	ID int32
}

func (o *SimpleOrganization) GetID() int32 { return o.ID }

// SimpleAccount is a simple implementation of AccountEntity.
// Use this if your domain entity doesn't already implement GetID().
type SimpleAccount struct {
	ID int32
}

func (a *SimpleAccount) GetID() int32 { return a.ID }
