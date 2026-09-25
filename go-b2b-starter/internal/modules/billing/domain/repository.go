package domain

import "context"

type OrganizationAdapter interface {
	GetExternalCustomerID(ctx context.Context, organizationID int32) (string, error)
}

type BillingProvider interface {
	Enabled() bool
	ProductID() string
	GetCustomerState(ctx context.Context, externalID string) (*CustomerState, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error)
}
