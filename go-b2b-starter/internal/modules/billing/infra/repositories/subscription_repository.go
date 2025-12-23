package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moasq/go-b2b-starter/internal/db/helpers"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
)

// subscriptionRepository implements domain.SubscriptionRepository using SQLC internally.
// SQLC types are never exposed outside this package.
type subscriptionRepository struct {
	store sqlc.Store
}

// NewSubscriptionRepository creates a new SubscriptionRepository implementation.
func NewSubscriptionRepository(store sqlc.Store) domain.SubscriptionRepository {
	return &subscriptionRepository{store: store}
}

func (r *subscriptionRepository) GetSubscriptionByOrgID(ctx context.Context, organizationID int32) (*domain.Subscription, error) {
	result, err := r.store.GetSubscriptionByOrgID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return r.mapToDomainSubscription(&result), nil
}

func (r *subscriptionRepository) UpsertSubscription(ctx context.Context, subscription *domain.Subscription) (*domain.Subscription, error) {
	// Marshal metadata to JSONB
	metadataJSON, err := json.Marshal(subscription.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := sqlc.UpsertSubscriptionParams{
		OrganizationID:     subscription.OrganizationID,
		ExternalCustomerID: subscription.ExternalCustomerID,
		SubscriptionID:     subscription.SubscriptionID,
		SubscriptionStatus: subscription.SubscriptionStatus,
		ProductID:          subscription.ProductID,
		ProductName:        helpers.ToPgText(subscription.ProductName),
		PlanName:           helpers.ToPgText(subscription.PlanName),
		CurrentPeriodStart: toPgTimestamp(subscription.CurrentPeriodStart),
		CurrentPeriodEnd:   toPgTimestamp(subscription.CurrentPeriodEnd),
		CancelAtPeriodEnd:  helpers.ToPgBool(subscription.CancelAtPeriodEnd),
		CanceledAt:         toPgTimestampPtr(subscription.CanceledAt),
		Metadata:           metadataJSON,
	}

	result, err := r.store.UpsertSubscription(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	return r.mapToDomainSubscription(&result), nil
}

func (r *subscriptionRepository) DeleteSubscription(ctx context.Context, organizationID int32) error {
	if err := r.store.DeleteSubscription(ctx, organizationID); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (r *subscriptionRepository) GetQuotaByOrgID(ctx context.Context, organizationID int32) (*domain.QuotaTracking, error) {
	result, err := r.store.GetQuotaByOrgID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrQuotaNotFound
		}
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	return r.mapToDomainQuota(&result), nil
}

func (r *subscriptionRepository) UpsertQuota(ctx context.Context, quota *domain.QuotaTracking) (*domain.QuotaTracking, error) {
	params := sqlc.UpsertQuotaParams{
		OrganizationID: quota.OrganizationID,
		InvoiceCount:   quota.InvoiceCount,
		MaxSeats:       helpers.ToPgInt4(quota.MaxSeats),
		PeriodStart:    toPgTimestamp(quota.PeriodStart),
		PeriodEnd:      toPgTimestamp(quota.PeriodEnd),
	}

	result, err := r.store.UpsertQuota(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert quota: %w", err)
	}

	return r.mapToDomainQuota(&result), nil
}

func (r *subscriptionRepository) DecrementInvoiceCount(ctx context.Context, organizationID int32) (*domain.QuotaTracking, error) {
	result, err := r.store.DecrementInvoiceCount(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrement invoice count: %w", err)
	}

	return r.mapToDomainQuota(&result), nil
}

func (r *subscriptionRepository) GetQuotaStatus(ctx context.Context, organizationID int32) (*domain.QuotaStatus, error) {
	result, err := r.store.GetQuotaStatus(ctx, organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get quota status: %w", err)
	}

	return r.mapToDomainQuotaStatus(&result), nil
}

// Mapping functions

func (r *subscriptionRepository) mapToDomainSubscription(s *sqlc.SubscriptionBillingSubscription) *domain.Subscription {
	var metadata map[string]any
	if len(s.Metadata) > 0 {
		json.Unmarshal(s.Metadata, &metadata)
	}

	subscription := &domain.Subscription{
		ID:                 s.ID,
		OrganizationID:     s.OrganizationID,
		ExternalCustomerID: s.ExternalCustomerID,
		SubscriptionID:     s.SubscriptionID,
		SubscriptionStatus: s.SubscriptionStatus,
		ProductID:          s.ProductID,
		ProductName:        helpers.FromPgText(s.ProductName),
		PlanName:           helpers.FromPgText(s.PlanName),
		CurrentPeriodStart: s.CurrentPeriodStart.Time,
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Time,
		Metadata:           metadata,
		CreatedAt:          s.CreatedAt.Time,
		UpdatedAt:          s.UpdatedAt.Time,
	}

	// Handle nullable fields
	if s.CancelAtPeriodEnd.Valid {
		subscription.CancelAtPeriodEnd = s.CancelAtPeriodEnd.Bool
	}
	if s.CanceledAt.Valid {
		subscription.CanceledAt = &s.CanceledAt.Time
	}

	return subscription
}

func (r *subscriptionRepository) mapToDomainQuota(q *sqlc.SubscriptionBillingQuotaTracking) *domain.QuotaTracking {
	quota := &domain.QuotaTracking{
		ID:             q.ID,
		OrganizationID: q.OrganizationID,
		InvoiceCount:   q.InvoiceCount,
		MaxSeats:       helpers.FromPgInt4(q.MaxSeats),
		PeriodStart:    q.PeriodStart.Time,
		PeriodEnd:      q.PeriodEnd.Time,
		CreatedAt:      q.CreatedAt.Time,
		UpdatedAt:      q.UpdatedAt.Time,
	}

	// Handle nullable LastSyncedAt
	if q.LastSyncedAt.Valid {
		quota.LastSyncedAt = &q.LastSyncedAt.Time
	}

	return quota
}

func (r *subscriptionRepository) mapToDomainQuotaStatus(qs *sqlc.GetQuotaStatusRow) *domain.QuotaStatus {
	status := &domain.QuotaStatus{
		SubscriptionStatus: qs.SubscriptionStatus,
		CurrentPeriodStart: qs.CurrentPeriodStart.Time,
		CurrentPeriodEnd:   qs.CurrentPeriodEnd.Time,
		InvoiceCount:       qs.InvoiceCount,
		CanProcessInvoice:  qs.CanProcessInvoice,
	}

	// Handle nullable fields
	if qs.CancelAtPeriodEnd.Valid {
		status.CancelAtPeriodEnd = qs.CancelAtPeriodEnd.Bool
	}
	if qs.MaxSeats.Valid {
		status.MaxSeats = qs.MaxSeats.Int32
	}

	return status
}

// Helper functions for timestamp conversion

func toPgTimestamp(t time.Time) pgtype.Timestamp {
	if t.IsZero() {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: t, Valid: true}
}

func toPgTimestampPtr(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}
