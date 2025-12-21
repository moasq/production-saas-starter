package adapters

import (
	"context"

	db "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// SubscriptionStore provides database operations for subscription billing
type SubscriptionStore interface {
	// Subscription operations
	GetSubscriptionByOrgID(ctx context.Context, organizationID int32) (db.SubscriptionBillingSubscription, error)
	GetSubscriptionBySubscriptionID(ctx context.Context, subscriptionID string) (db.SubscriptionBillingSubscription, error)
	UpsertSubscription(ctx context.Context, arg db.UpsertSubscriptionParams) (db.SubscriptionBillingSubscription, error)
	DeleteSubscription(ctx context.Context, organizationID int32) error
	ListActiveSubscriptions(ctx context.Context) ([]db.SubscriptionBillingSubscription, error)

	// Quota operations
	GetQuotaByOrgID(ctx context.Context, organizationID int32) (db.SubscriptionBillingQuotaTracking, error)
	UpsertQuota(ctx context.Context, arg db.UpsertQuotaParams) (db.SubscriptionBillingQuotaTracking, error)
	DecrementInvoiceCount(ctx context.Context, organizationID int32) (db.SubscriptionBillingQuotaTracking, error)
	ResetQuotaForPeriod(ctx context.Context, arg db.ResetQuotaForPeriodParams) (db.SubscriptionBillingQuotaTracking, error)

	// Combined operations
	GetQuotaStatus(ctx context.Context, organizationID int32) (db.GetQuotaStatusRow, error)
	ListQuotasNearLimit(ctx context.Context, threshold int32) ([]db.ListQuotasNearLimitRow, error)
}
