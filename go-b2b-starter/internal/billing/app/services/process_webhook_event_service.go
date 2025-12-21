package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
)

const invoicesProcessedMeterSlug = "invoice.processed"

func (s *billingService) ProcessWebhookEvent(ctx context.Context, eventType string, payload map[string]any) error {
	s.logger.Info("Processing webhook event", map[string]any{
		"event_type":   eventType,
		"payload_keys": mapKeys(payload),
	})

	// Update subscription based on event type
	switch eventType {
	case "subscription.created", "subscription.updated":
		eventData, err := s.parseSubscriptionWebhookPayload(payload)
		if err != nil {
			return fmt.Errorf("failed to parse subscription webhook payload: %w", err)
		}
		return s.handleSubscriptionUpsert(ctx, eventData)
	case "subscription.canceled":
		eventData, err := s.parseSubscriptionWebhookPayload(payload)
		if err != nil {
			return fmt.Errorf("failed to parse subscription webhook payload: %w", err)
		}
		return s.handleSubscriptionCanceled(ctx, eventData)
	case "customer.updated":
		eventData, err := s.parseSubscriptionWebhookPayload(payload)
		if err != nil {
			return fmt.Errorf("failed to parse subscription webhook payload: %w", err)
		}
		return s.handleCustomerUpdated(ctx, eventData)
	case "meter.grant.updated", "meter.grant.created", "entitlement.grant.updated":
		if err := s.handleMeterGrantEvent(ctx, payload); err != nil {
			return fmt.Errorf("failed to handle meter grant webhook: %w", err)
		}
		return nil
	default:
		s.logger.Warn("Unhandled webhook event type", map[string]any{
			"event_type": eventType,
		})
		return nil // Don't fail on unknown events
	}
}

func (s *billingService) parseSubscriptionWebhookPayload(payload map[string]any) (*domain.SubscriptionEventData, error) {
	normalized := normalizePolarObject(payload)
	if normalized == nil {
		return nil, fmt.Errorf("webhook payload missing subscription object")
	}

	data := &domain.SubscriptionEventData{}

	if subID, ok := normalized["id"].(string); ok {
		data.SubscriptionID = subID
	} else if subID, ok := normalized["subscription_id"].(string); ok {
		data.SubscriptionID = subID
	}

	if status, ok := normalized["status"].(string); ok {
		data.Status = status
	}

	if t, ok := parseISOTime(normalized["current_period_start"]); ok {
		data.CurrentPeriodStart = t
	} else if t, ok := parseISOTime(normalized["current_period_start_at"]); ok {
		data.CurrentPeriodStart = t
	}

	if t, ok := parseISOTime(normalized["current_period_end"]); ok {
		data.CurrentPeriodEnd = t
	} else if t, ok := parseISOTime(normalized["current_period_end_at"]); ok {
		data.CurrentPeriodEnd = t
	}

	if value, exists := normalized["cancel_at_period_end"]; exists {
		if v, ok := toBool(value); ok {
			data.CancelAtPeriodEnd = v
		}
	}

	if value, exists := normalized["canceled_at"]; exists {
		if t, ok := parseISOTime(value); ok {
			data.CanceledAt = &t
		}
	}

	product := extractProductMap(normalized)
	if product == nil {
		product = extractProductMap(payload)
	}

	if product != nil {
		if productID, ok := product["id"].(string); ok && data.ProductID == "" {
			data.ProductID = productID
		}
		if productName, ok := product["name"].(string); ok && data.ProductName == "" {
			data.ProductName = productName
		}
		if metadata := stringMapFrom(product["metadata"]); len(metadata) > 0 {
			data.ProductMetadata = metadata
		}
	}

	if data.ProductID == "" {
		if productID, ok := normalized["product_id"].(string); ok {
			data.ProductID = productID
		} else if productID, ok := payload["product_id"].(string); ok {
			data.ProductID = productID
		}
	}

	if data.ProductName == "" {
		if productName, ok := normalized["product_name"].(string); ok {
			data.ProductName = productName
		}
	}

	if len(data.ProductMetadata) == 0 {
		if metadata := stringMapFrom(normalized["product_metadata"]); len(metadata) > 0 {
			data.ProductMetadata = metadata
		} else if metadata := stringMapFrom(payload["product_metadata"]); len(metadata) > 0 {
			data.ProductMetadata = metadata
		}
	}

	if product != nil {
		if invoiceCount := extractInvoiceCountFromProduct(product); invoiceCount != "" {
			if data.ProductMetadata == nil {
				data.ProductMetadata = make(map[string]string)
			}
			if existing, ok := data.ProductMetadata["invoice_count"]; !ok || existing == "" {
				data.ProductMetadata["invoice_count"] = invoiceCount
			}
		}
	}

	if metadata := stringMapFrom(normalized["metadata"]); len(metadata) > 0 {
		data.CustomerMetadata = metadata
	}

	if len(data.CustomerMetadata) == 0 {
		if customer, ok := normalized["customer"].(map[string]any); ok {
			if metadata := stringMapFrom(customer["metadata"]); len(metadata) > 0 {
				data.CustomerMetadata = metadata
			}

			if data.ExternalCustomerID == "" {
				if externalID, ok := customer["external_id"].(string); ok && externalID != "" {
					data.ExternalCustomerID = externalID
				} else if externalID, ok := customer["id"].(string); ok && externalID != "" {
					data.ExternalCustomerID = externalID
				}
			}
		}
	}

	if len(data.CustomerMetadata) == 0 {
		if metadata := stringMapFrom(payload["metadata"]); len(metadata) > 0 {
			data.CustomerMetadata = metadata
		}
	}

	if data.ExternalCustomerID == "" {
		if externalID, ok := normalized["customer_external_id"].(string); ok && externalID != "" {
			data.ExternalCustomerID = externalID
		} else if externalID, ok := normalized["external_customer_id"].(string); ok && externalID != "" {
			data.ExternalCustomerID = externalID
		} else if externalID, ok := payload["customer_external_id"].(string); ok && externalID != "" {
			data.ExternalCustomerID = externalID
		}
	}

	if data.ExternalCustomerID == "" && len(data.CustomerMetadata) > 0 {
		if externalID, ok := data.CustomerMetadata["organization_id"]; ok && externalID != "" {
			data.ExternalCustomerID = externalID
		} else if externalID, ok := data.CustomerMetadata["external_customer_id"]; ok && externalID != "" {
			data.ExternalCustomerID = externalID
		}
	}

	if data.ExternalCustomerID == "" {
		s.logger.Warn("Subscription webhook payload missing external customer identifier", map[string]any{
			"payload_keys": mapKeys(normalized),
		})
		return nil, fmt.Errorf("webhook payload missing organization_id")
	}

	s.logger.Info("Parsed subscription webhook payload", map[string]any{
		"subscription_id":        data.SubscriptionID,
		"external_customer_id":   data.ExternalCustomerID,
		"status":                 data.Status,
		"product_id":             data.ProductID,
		"product_metadata_keys":  len(data.ProductMetadata),
		"customer_metadata_keys": len(data.CustomerMetadata),
	})

	return data, nil
}

func (s *billingService) handleSubscriptionUpsert(ctx context.Context, eventData *domain.SubscriptionEventData) error {
	// Step 1: Map Polar organization_id to internal organization ID
	organizationID, err := s.orgAdapter.GetOrganizationIDByStytchOrgID(ctx, eventData.ExternalCustomerID)
	if err != nil {
		return fmt.Errorf("failed to map organization: %w", err)
	}

	s.logger.Info("Mapped organization", map[string]any{
		"external_customer_id": eventData.ExternalCustomerID,
		"organization_id":      organizationID,
	})

	// Step 2: Parse quota limits from product metadata (remaining invoices)
	var invoiceCount int32 = 0
	if val, ok := eventData.ProductMetadata["invoice_count"]; ok {
		if count, err := strconv.ParseInt(val, 10, 32); err == nil {
			invoiceCount = int32(count)
		} else {
			s.logger.Warn("Failed to parse invoice_count from product metadata", map[string]any{
				"value": val,
				"error": err.Error(),
			})
		}
	} else {
		s.logger.Warn("invoice_count not found in product metadata", map[string]any{
			"product_metadata": eventData.ProductMetadata,
		})
	}

	var maxSeats int32 = 0
	if val, ok := eventData.ProductMetadata["max_seats"]; ok {
		if count, err := strconv.ParseInt(val, 10, 32); err == nil {
			maxSeats = int32(count)
		}
	}

	s.logger.Info("Parsed quota limits from metadata", map[string]any{
		"invoice_count": invoiceCount,
		"max_seats":     maxSeats,
	})

	// Step 4: Create subscription domain object
	subscription := &domain.Subscription{
		OrganizationID:     organizationID,
		ExternalCustomerID: eventData.ExternalCustomerID,
		SubscriptionID:     eventData.SubscriptionID,
		SubscriptionStatus: eventData.Status,
		ProductID:          eventData.ProductID,
		ProductName:        eventData.ProductName,
		CurrentPeriodStart: eventData.CurrentPeriodStart,
		CurrentPeriodEnd:   eventData.CurrentPeriodEnd,
		CancelAtPeriodEnd:  eventData.CancelAtPeriodEnd,
		CanceledAt:         eventData.CanceledAt,
	}

	// Step 5: Upsert subscription to database
	_, err = s.repo.UpsertSubscription(ctx, subscription)
	if err != nil {
		return fmt.Errorf("failed to upsert subscription: %w", err)
	}

	s.logger.Info("Upserted subscription", map[string]any{
		"organization_id": organizationID,
		"subscription_id": eventData.SubscriptionID,
		"status":          eventData.Status,
	})

	// Step 6: Create quota tracking domain object
	now := time.Now()
	quota := &domain.QuotaTracking{
		OrganizationID: organizationID,
		InvoiceCount:   invoiceCount,
		MaxSeats:       maxSeats,
		PeriodStart:    eventData.CurrentPeriodStart,
		PeriodEnd:      eventData.CurrentPeriodEnd,
		LastSyncedAt:   &now,
	}

	// Step 7: Upsert quota tracking to database
	_, err = s.repo.UpsertQuota(ctx, quota)
	if err != nil {
		return fmt.Errorf("failed to upsert quota: %w", err)
	}

	s.logger.Info("Upserted quota tracking", map[string]any{
		"organization_id": organizationID,
		"invoice_count":   invoiceCount,
		"max_seats":       maxSeats,
	})

	return nil
}

func (s *billingService) handleSubscriptionCanceled(ctx context.Context, eventData *domain.SubscriptionEventData) error {
	// Step 1: Map Polar organization_id to internal organization ID
	organizationID, err := s.orgAdapter.GetOrganizationIDByStytchOrgID(ctx, eventData.ExternalCustomerID)
	if err != nil {
		return fmt.Errorf("failed to map organization: %w", err)
	}

	s.logger.Info("Processing subscription cancellation", map[string]any{
		"organization_id": organizationID,
		"subscription_id": eventData.SubscriptionID,
	})

	// Step 2: Create subscription object with canceled status
	now := time.Now()
	subscription := &domain.Subscription{
		OrganizationID:     organizationID,
		ExternalCustomerID: eventData.ExternalCustomerID,
		SubscriptionID:     eventData.SubscriptionID,
		SubscriptionStatus: "canceled",
		ProductID:          eventData.ProductID,
		ProductName:        eventData.ProductName,
		CurrentPeriodStart: eventData.CurrentPeriodStart,
		CurrentPeriodEnd:   eventData.CurrentPeriodEnd,
		CancelAtPeriodEnd:  false, // Already canceled
		CanceledAt:         &now,
	}

	// If webhook includes canceled_at timestamp, use it
	if eventData.CanceledAt != nil {
		subscription.CanceledAt = eventData.CanceledAt
	}

	// Step 3: Upsert subscription with canceled status
	_, err = s.repo.UpsertSubscription(ctx, subscription)
	if err != nil {
		return fmt.Errorf("failed to update subscription to canceled: %w", err)
	}

	s.logger.Info("Subscription marked as canceled", map[string]any{
		"organization_id": organizationID,
		"subscription_id": eventData.SubscriptionID,
		"canceled_at":     subscription.CanceledAt,
	})

	return nil
}

func (s *billingService) handleCustomerUpdated(ctx context.Context, eventData *domain.SubscriptionEventData) error {
	// Step 1: Map Polar organization_id to internal organization ID
	organizationID, err := s.orgAdapter.GetOrganizationIDByStytchOrgID(ctx, eventData.ExternalCustomerID)
	if err != nil {
		return fmt.Errorf("failed to map organization: %w", err)
	}

	s.logger.Info("Processing customer update", map[string]any{
		"organization_id": organizationID,
		"metadata_keys":   len(eventData.CustomerMetadata),
	})

	// Step 2: Parse invoice count from customer metadata (remaining count)
	var invoiceCount int32 = 0
	if val, ok := eventData.CustomerMetadata["invoice_count"]; ok {
		if count, err := strconv.ParseInt(val, 10, 32); err == nil {
			invoiceCount = int32(count)
		}
	}

	// Step 3: Get existing quota to preserve other fields
	existingQuota, err := s.repo.GetQuotaByOrgID(ctx, organizationID)
	if err != nil {
		// If no quota exists, create a minimal one with just the invoice count
		s.logger.Warn("No existing quota found, creating new quota entry", map[string]any{
			"organization_id": organizationID,
		})

		now := time.Now()
		quota := &domain.QuotaTracking{
			OrganizationID: organizationID,
			InvoiceCount:   invoiceCount,
			MaxSeats:       0,
			PeriodStart:    now,
			PeriodEnd:      now,
			LastSyncedAt:   &now,
		}

		_, err = s.repo.UpsertQuota(ctx, quota)
		if err != nil {
			return fmt.Errorf("failed to create quota: %w", err)
		}

		s.logger.Info("Created new quota with invoice count", map[string]any{
			"organization_id": organizationID,
			"invoice_count":   invoiceCount,
		})

		return nil
	}

	// Step 4: Update existing quota with new invoice count
	now := time.Now()
	_ = existingQuota.InvoiceCount
	existingQuota.InvoiceCount = invoiceCount
	existingQuota.LastSyncedAt = &now

	_, err = s.repo.UpsertQuota(ctx, existingQuota)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	s.logger.Info("Updated quota with invoice count from customer metadata", map[string]any{
		"organization_id": organizationID,
		"invoice_count":   invoiceCount,
	})

	return nil
}

func (s *billingService) handleMeterGrantEvent(ctx context.Context, payload map[string]any) error {
	eventData, err := s.parseMeterGrantPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to parse meter grant payload: %w", err)
	}

	if !strings.EqualFold(eventData.MeterSlug, invoicesProcessedMeterSlug) {
		s.logger.Info("Ignoring meter grant event for unrelated meter", map[string]any{
			"meter_slug": eventData.MeterSlug,
		})
		return nil
	}

	organizationID, err := s.orgAdapter.GetOrganizationIDByStytchOrgID(ctx, eventData.ExternalCustomerID)
	if err != nil {
		return fmt.Errorf("failed to map organization for meter grant: %w", err)
	}

	now := time.Now()
	quota, err := s.repo.GetQuotaByOrgID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, domain.ErrQuotaNotFound) {
			newQuota := &domain.QuotaTracking{
				OrganizationID: organizationID,
				InvoiceCount:   eventData.AvailableCredits,
				MaxSeats:       0,
				PeriodStart:    now,
				PeriodEnd:      now,
				LastSyncedAt:   &now,
			}

			if _, err := s.repo.UpsertQuota(ctx, newQuota); err != nil {
				return fmt.Errorf("failed to create quota from meter grant: %w", err)
			}

			s.logger.Info("Created quota from meter grant event", map[string]any{
				"organization_id": organizationID,
				"meter_slug":      eventData.MeterSlug,
				"invoice_count":   eventData.AvailableCredits,
			})

			return nil
		}

		return fmt.Errorf("failed to get quota for meter grant: %w", err)
	}

	previous := quota.InvoiceCount
	quota.InvoiceCount = eventData.AvailableCredits
	quota.LastSyncedAt = &now

	if _, err := s.repo.UpsertQuota(ctx, quota); err != nil {
		return fmt.Errorf("failed to update quota from meter grant: %w", err)
	}

	s.logger.Info("Updated quota from meter grant event", map[string]any{
		"organization_id": organizationID,
		"meter_slug":      eventData.MeterSlug,
		"invoice_count":   quota.InvoiceCount,
		"previous_count":  previous,
	})

	return nil
}

func (s *billingService) parseMeterGrantPayload(payload map[string]any) (*domain.MeterGrantEventData, error) {
	normalized := normalizePolarObject(payload)
	if normalized == nil {
		return nil, fmt.Errorf("meter grant payload missing object")
	}

	data := &domain.MeterGrantEventData{}

	if slug, ok := toString(normalized["meter_slug"]); ok {
		data.MeterSlug = strings.TrimSpace(slug)
	}
	if data.MeterSlug == "" {
		if slug, ok := toString(normalized["slug"]); ok {
			data.MeterSlug = strings.TrimSpace(slug)
		}
	}
	if data.MeterSlug == "" {
		if meter, ok := normalized["meter"].(map[string]any); ok {
			if slug, ok := toString(meter["slug"]); ok {
				data.MeterSlug = strings.TrimSpace(slug)
			} else if slug, ok := toString(meter["meter_slug"]); ok {
				data.MeterSlug = strings.TrimSpace(slug)
			} else if slug, ok := toString(meter["name"]); ok {
				data.MeterSlug = strings.TrimSpace(slug)
			}
		}
	}

	if externalID, ok := toString(normalized["external_customer_id"]); ok && strings.TrimSpace(externalID) != "" {
		data.ExternalCustomerID = strings.TrimSpace(externalID)
	}
	if data.ExternalCustomerID == "" {
		if externalID, ok := toString(normalized["customer_external_id"]); ok && strings.TrimSpace(externalID) != "" {
			data.ExternalCustomerID = strings.TrimSpace(externalID)
		}
	}
	if data.ExternalCustomerID == "" {
		if customer, ok := normalized["customer"].(map[string]any); ok {
			if externalID, ok := toString(customer["external_id"]); ok && strings.TrimSpace(externalID) != "" {
				data.ExternalCustomerID = strings.TrimSpace(externalID)
			} else if externalID, ok := toString(customer["id"]); ok && strings.TrimSpace(externalID) != "" {
				data.ExternalCustomerID = strings.TrimSpace(externalID)
			} else if metadata := stringMapFrom(customer["metadata"]); len(metadata) > 0 {
				if externalID := strings.TrimSpace(metadata["organization_id"]); externalID != "" {
					data.ExternalCustomerID = externalID
				}
			}
		}
	}
	if data.ExternalCustomerID == "" {
		if metadata := stringMapFrom(normalized["metadata"]); len(metadata) > 0 {
			if externalID := strings.TrimSpace(metadata["organization_id"]); externalID != "" {
				data.ExternalCustomerID = externalID
			}
		}
	}

	var (
		available  int32
		hasBalance bool
	)

	if balanceMap, ok := normalized["balance"].(map[string]any); ok {
		for _, key := range []string{"available", "remaining", "quantity", "value"} {
			if value, exists := balanceMap[key]; exists {
				if count, ok := toInt32(value); ok {
					available = count
					hasBalance = true
					break
				}
			}
		}
	}

	if !hasBalance {
		if creditBalance, ok := normalized["credit_balance"].(map[string]any); ok {
			for _, key := range []string{"available", "remaining", "quantity"} {
				if value, exists := creditBalance[key]; exists {
					if count, ok := toInt32(value); ok {
						available = count
						hasBalance = true
						break
					}
				}
			}
		}
	}

	if !hasBalance {
		for _, key := range []string{"available", "remaining", "balance", "quantity"} {
			if value, exists := normalized[key]; exists {
				if count, ok := toInt32(value); ok {
					available = count
					hasBalance = true
					break
				}
			}
		}
	}

	if !hasBalance {
		s.logger.Warn("Meter grant payload missing available balance", map[string]any{
			"payload_keys": mapKeys(normalized),
		})
		return nil, fmt.Errorf("meter grant payload missing available balance")
	}

	data.AvailableCredits = available

	if data.MeterSlug == "" {
		s.logger.Warn("Meter grant payload missing meter slug", map[string]any{
			"payload_keys": mapKeys(normalized),
		})
		return nil, fmt.Errorf("meter grant payload missing meter slug")
	}

	if data.ExternalCustomerID == "" {
		s.logger.Warn("Meter grant payload missing external customer identifier", map[string]any{
			"payload_keys": mapKeys(normalized),
		})
		return nil, fmt.Errorf("meter grant payload missing external customer id")
	}

	s.logger.Info("Parsed meter grant payload", map[string]any{
		"meter_slug":            data.MeterSlug,
		"external_customer_id":  data.ExternalCustomerID,
		"available_invoice_cnt": data.AvailableCredits,
	})

	return data, nil
}

func normalizePolarObject(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}

	if object, ok := payload["object"].(map[string]any); ok && len(object) > 0 {
		return object
	}

	if data, ok := payload["data"].(map[string]any); ok {
		if object, ok := data["object"].(map[string]any); ok && len(object) > 0 {
			return object
		}
	}

	if dataSlice, ok := payload["data"].([]any); ok && len(dataSlice) > 0 {
		for _, item := range dataSlice {
			if itemMap, ok := item.(map[string]any); ok {
				if object, ok := itemMap["object"].(map[string]any); ok && len(object) > 0 {
					return object
				}
			}
		}
	}

	return payload
}

func extractProductMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	if product, ok := input["product"].(map[string]any); ok {
		return product
	}

	if price, ok := input["price"].(map[string]any); ok {
		if product, ok := price["product"].(map[string]any); ok {
			return product
		}
	}

	if plan, ok := input["plan"].(map[string]any); ok {
		if product, ok := plan["product"].(map[string]any); ok {
			return product
		}
	}

	if itemsMap := firstMapFromSlice(input["items"]); itemsMap != nil {
		if product, ok := itemsMap["product"].(map[string]any); ok {
			return product
		}
		if price, ok := itemsMap["price"].(map[string]any); ok {
			if product, ok := price["product"].(map[string]any); ok {
				return product
			}
		}
	}

	return nil
}

func firstMapFromSlice(value any) map[string]any {
	items, ok := value.([]any)
	if !ok {
		return nil
	}

	for _, item := range items {
		if itemMap, ok := item.(map[string]any); ok {
			return itemMap
		}
	}

	return nil
}

func stringMapFrom(value any) map[string]string {
	source, ok := value.(map[string]any)
	if !ok || len(source) == 0 {
		return nil
	}

	result := toStringMap(source)
	if len(result) == 0 {
		return nil
	}

	return result
}

func toStringMap(input map[string]any) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		if str, ok := toString(value); ok {
			result[key] = str
		}
	}
	return result
}

func toString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	case bool:
		return strconv.FormatBool(v), true
	case int:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return "", false
		}
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint:
		if v > uint(math.MaxInt32) {
			return "", false
		}
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case float32:
		f := float64(v)
		if math.Mod(f, 1) == 0 {
			return strconv.FormatInt(int64(f), 10), true
		}
		return strconv.FormatFloat(f, 'f', -1, 32), true
	case float64:
		if math.Mod(v, 1) == 0 {
			return strconv.FormatInt(int64(v), 10), true
		}
		return strconv.FormatFloat(v, 'f', -1, 64), true
	default:
		return "", false
	}
}

func toInt32(value any) (int32, bool) {
	switch v := value.(type) {
	case int:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return 0, false
		}
		return int32(v), true
	case int8:
		return int32(v), true
	case int16:
		return int32(v), true
	case int32:
		return v, true
	case int64:
		if v > int64(math.MaxInt32) || v < int64(math.MinInt32) {
			return 0, false
		}
		return int32(v), true
	case uint:
		if v > uint(math.MaxInt32) {
			return 0, false
		}
		return int32(v), true
	case uint8:
		return int32(v), true
	case uint16:
		return int32(v), true
	case uint32:
		if v > uint32(math.MaxInt32) {
			return 0, false
		}
		return int32(v), true
	case uint64:
		if v > uint64(math.MaxInt32) {
			return 0, false
		}
		return int32(v), true
	case float32:
		f := float64(v)
		if math.Mod(f, 1) != 0 {
			return 0, false
		}
		if f > float64(math.MaxInt32) || f < float64(math.MinInt32) {
			return 0, false
		}
		return int32(f), true
	case float64:
		if math.Mod(v, 1) != 0 {
			return 0, false
		}
		if v > float64(math.MaxInt32) || v < float64(math.MinInt32) {
			return 0, false
		}
		return int32(v), true
	case string:
		if strings.TrimSpace(v) == "" {
			return 0, false
		}
		if strings.Contains(v, ".") {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return 0, false
			}
			if math.Mod(f, 1) != 0 {
				return 0, false
			}
			if f > float64(math.MaxInt32) || f < float64(math.MinInt32) {
				return 0, false
			}
			return int32(f), true
		}
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return int32(i), true
	default:
		return 0, false
	}
}

func parseISOTime(value any) (time.Time, bool) {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return time.Time{}, false
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	case time.Time:
		return v, true
	default:
		return time.Time{}, false
	}
}

func toBool(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if strings.TrimSpace(v) == "" {
			return false, false
		}
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return false, false
		}
		return parsed, true
	case int:
		return v != 0, true
	case int32:
		return v != 0, true
	case int64:
		return v != 0, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}

func mapKeys(m map[string]any) []string {
	if m == nil {
		return nil
	}

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func extractInvoiceCountFromProduct(product map[string]any) string {
	if product == nil {
		return ""
	}

	if metadata := stringMapFrom(product["metadata"]); len(metadata) > 0 {
		if value := strings.TrimSpace(metadata["invoice_count"]); value != "" {
			return value
		}
	}

	benefits, ok := product["benefits"].([]any)
	if !ok || len(benefits) == 0 {
		return ""
	}

	for _, item := range benefits {
		benefit, ok := item.(map[string]any)
		if !ok {
			continue
		}

		benefitType, _ := toString(benefit["type"])
		if !strings.EqualFold(strings.TrimSpace(benefitType), "meter_credit") {
			continue
		}

		if properties, ok := benefit["properties"].(map[string]any); ok {
			if count, ok := toInt32(properties["units"]); ok && count > 0 {
				return strconv.FormatInt(int64(count), 10)
			}
		}

		if metadata := stringMapFrom(benefit["metadata"]); len(metadata) > 0 {
			if value := strings.TrimSpace(metadata["units"]); value != "" {
				return value
			}
		}
	}

	return ""
}
