package adapterimpl

import (
	"context"

	sqlc "github.com/moasq/go-b2b-starter/pkg/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/pkg/db/adapters"
)

// subscriptionStore implements adapters.SubscriptionStore
type subscriptionStore struct {
	store sqlc.Store
}

func NewSubscriptionStore(store sqlc.Store) adapters.SubscriptionStore {
	return &subscriptionStore{store: store}
}

// Subscription operations

func (s *subscriptionStore) GetSubscriptionByOrgID(ctx context.Context, organizationID int32) (sqlc.SubscriptionBillingSubscription, error) {
	return s.store.GetSubscriptionByOrgID(ctx, organizationID)
}

func (s *subscriptionStore) GetSubscriptionBySubscriptionID(ctx context.Context, subscriptionID string) (sqlc.SubscriptionBillingSubscription, error) {
	return s.store.GetSubscriptionBySubscriptionID(ctx, subscriptionID)
}

func (s *subscriptionStore) UpsertSubscription(ctx context.Context, arg sqlc.UpsertSubscriptionParams) (sqlc.SubscriptionBillingSubscription, error) {
	return s.store.UpsertSubscription(ctx, arg)
}

func (s *subscriptionStore) DeleteSubscription(ctx context.Context, organizationID int32) error {
	return s.store.DeleteSubscription(ctx, organizationID)
}

func (s *subscriptionStore) ListActiveSubscriptions(ctx context.Context) ([]sqlc.SubscriptionBillingSubscription, error) {
	return s.store.ListActiveSubscriptions(ctx)
}

// Quota operations

func (s *subscriptionStore) GetQuotaByOrgID(ctx context.Context, organizationID int32) (sqlc.SubscriptionBillingQuotaTracking, error) {
	return s.store.GetQuotaByOrgID(ctx, organizationID)
}

func (s *subscriptionStore) UpsertQuota(ctx context.Context, arg sqlc.UpsertQuotaParams) (sqlc.SubscriptionBillingQuotaTracking, error) {
	return s.store.UpsertQuota(ctx, arg)
}

func (s *subscriptionStore) DecrementInvoiceCount(ctx context.Context, organizationID int32) (sqlc.SubscriptionBillingQuotaTracking, error) {
	return s.store.DecrementInvoiceCount(ctx, organizationID)
}

func (s *subscriptionStore) ResetQuotaForPeriod(ctx context.Context, arg sqlc.ResetQuotaForPeriodParams) (sqlc.SubscriptionBillingQuotaTracking, error) {
	return s.store.ResetQuotaForPeriod(ctx, arg)
}

// Combined operations

func (s *subscriptionStore) GetQuotaStatus(ctx context.Context, organizationID int32) (sqlc.GetQuotaStatusRow, error) {
	return s.store.GetQuotaStatus(ctx, organizationID)
}

func (s *subscriptionStore) ListQuotasNearLimit(ctx context.Context, threshold int32) ([]sqlc.ListQuotasNearLimitRow, error) {
	return s.store.ListQuotasNearLimit(ctx, threshold)
}
