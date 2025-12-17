package polar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/moasq/go-b2b-starter/app/billing/domain"
	polarpkg "github.com/moasq/go-b2b-starter/pkg/polar"
)

type polarAdapter struct {
	client *polarpkg.Client
}

func NewPolarAdapter(client *polarpkg.Client) *polarAdapter {
	return &polarAdapter{
		client: client,
	}
}

func (p *polarAdapter) GetSubscription(ctx context.Context, externalCustomerID string) (*domain.Subscription, error) {
	// Call Polar API to get subscription by customer external ID
	endpoint := fmt.Sprintf("/v1/subscriptions?customer_external_id=%s", externalCustomerID)

	resp, err := p.client.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to call Polar API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("polar API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Items []struct {
			ID                 string `json:"id"`
			CustomerID         string `json:"customer_id"`
			ProductID          string `json:"product_id"`
			Status             string `json:"status"`
			CurrentPeriodStart string `json:"current_period_start"`
			CurrentPeriodEnd   string `json:"current_period_end"`
			CanceledAt         *string `json:"canceled_at"`
			Customer           struct {
				ID       string            `json:"id"`
				Metadata map[string]string `json:"metadata"`
			} `json:"customer"`
			Product struct {
				ID       string            `json:"id"`
				Name     string            `json:"name"`
				Metadata map[string]string `json:"metadata"`
			} `json:"product"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, domain.ErrSubscriptionNotFound
	}

	polarSub := result.Items[0]

	// Parse timestamps
	currentPeriodStart, _ := parseTime(polarSub.CurrentPeriodStart)
	currentPeriodEnd, _ := parseTime(polarSub.CurrentPeriodEnd)

	var canceledAt *time.Time
	if polarSub.CanceledAt != nil {
		t, _ := parseTime(*polarSub.CanceledAt)
		canceledAt = &t
	}

	// Parse quota limit from product metadata
	invoiceCountMax := int32(0)
	if val, ok := polarSub.Product.Metadata["invoice_count"]; ok {
		if count, err := strconv.ParseInt(val, 10, 32); err == nil {
			invoiceCountMax = int32(count)
		}
	}

	// Console log for Polar API response
	fmt.Printf("🌐 POLAR API SYNC - Customer: %s | Subscription: %s | Invoice Count: %d | Status: %s | Product: %s\n",
		externalCustomerID, polarSub.ID, invoiceCountMax, polarSub.Status, polarSub.Product.Name)

	// Create domain subscription (organizationID will be set by caller)
	subscription := &domain.Subscription{
		ExternalCustomerID: externalCustomerID,
		SubscriptionID:     polarSub.ID,
		SubscriptionStatus: polarSub.Status,
		ProductID:          polarSub.ProductID,
		ProductName:        polarSub.Product.Name,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		CanceledAt:         canceledAt,
		Metadata: map[string]any{
			"invoice_count_max":    invoiceCountMax,
			"product_metadata":     polarSub.Product.Metadata,
			"customer_metadata":    polarSub.Customer.Metadata,
		},
	}

	return subscription, nil
}

// GetCheckoutSession retrieves checkout session details from Polar
func (p *polarAdapter) GetCheckoutSession(ctx context.Context, sessionID string) (*domain.CheckoutSessionResponse, error) {
	// Call Polar API to get checkout session details
	endpoint := fmt.Sprintf("/v1/checkouts/custom/%s", sessionID)

	resp, err := p.client.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to call Polar checkout API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("checkout session not found: %s", sessionID)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("polar checkout API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Polar returns customer_external_id at root level
	var result struct {
		ID                 string `json:"id"`
		Status             string `json:"status"`
		Amount             int64  `json:"amount"`
		CustomerExternalID string `json:"customer_external_id"` // The Stytch org ID we passed during checkout
		CustomerID         string `json:"customer_id"`          // Polar internal customer ID
		Product            struct {
			ID string `json:"id"`
		} `json:"product"`
		Customer struct {
			ID         string `json:"id"`
			ExternalID string `json:"external_id"` // Also available in nested customer object
		} `json:"customer"`
		Subscription struct {
			ID string `json:"id"`
		} `json:"subscription"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode checkout response: %w", err)
	}

	// Parse timestamp
	createdAt, _ := parseTime(result.CreatedAt)

	// Resolve external customer ID - try multiple fields
	externalCustomerID := result.CustomerExternalID
	if externalCustomerID == "" {
		externalCustomerID = result.Customer.ExternalID
	}

	// Console log for checkout session retrieval
	fmt.Printf("🔍 POLAR CHECKOUT SESSION - ID: %s | Status: %s | ExternalCustomerID: %s | CustomerID: %s | Subscription: %s\n",
		result.ID, result.Status, externalCustomerID, result.CustomerID, result.Subscription.ID)

	// Create domain checkout session response
	checkoutSession := &domain.CheckoutSessionResponse{
		ID:             result.ID,
		Status:         result.Status,
		CustomerID:     externalCustomerID, // Use external customer ID (Stytch org ID)
		SubscriptionID: result.Subscription.ID,
		ProductID:      result.Product.ID,
		Amount:         result.Amount,
		CreatedAt:      createdAt,
	}

	return checkoutSession, nil
}

// GetCheckoutSessionWithPolling retrieves checkout session with polling and retry logic
// Polls every 2 seconds for up to 10 seconds (5 attempts total)
// Continues polling when status is "pending" or on transient errors
// Returns immediately on "succeeded" status or non-retryable errors
func (p *polarAdapter) GetCheckoutSessionWithPolling(ctx context.Context, sessionID string) (*domain.CheckoutSessionResponse, error) {
	const (
		pollInterval = 2 * time.Second  // Poll every 2 seconds
		maxDuration  = 10 * time.Second // Total timeout: 10 seconds
	)

	deadline := time.Now().Add(maxDuration)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// First attempt (immediate)
	session, err := p.GetCheckoutSession(ctx, sessionID)
	if err == nil && session.Status == "succeeded" {
		return session, nil
	}

	// Log initial status
	if err == nil {
		fmt.Printf("🔄 [Polar Polling] Initial status: %s, will poll for %v\n", session.Status, maxDuration)
	} else if !isRetryableError(err) {
		// Non-retryable error (e.g., 404) - fail immediately
		return nil, err
	} else {
		fmt.Printf("⚠️  [Polar Polling] Initial attempt failed: %v, will retry\n", err)
	}

	// Polling loop
	attemptCount := 1
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			attemptCount++
			session, err := p.GetCheckoutSession(ctx, sessionID)

			if err == nil {
				fmt.Printf("🔄 [Polar Polling] Attempt %d: status=%s\n", attemptCount, session.Status)
				if session.Status == "succeeded" {
					fmt.Printf("✅ [Polar Polling] Success after %d attempts\n", attemptCount)
					return session, nil
				}
				// Continue polling for "pending", "processing", etc.
				continue
			}

			// Check if error is retryable
			if !isRetryableError(err) {
				fmt.Printf("❌ [Polar Polling] Non-retryable error: %v\n", err)
				return nil, err
			}

			fmt.Printf("⚠️  [Polar Polling] Attempt %d failed (retryable): %v\n", attemptCount, err)
		}
	}

	// Timeout reached - get last known status
	lastStatus := "unknown"
	if session != nil {
		lastStatus = session.Status
	}
	fmt.Printf("⏱️  [Polar Polling] Timeout after %d attempts (10s), last status: %s\n", attemptCount, lastStatus)
	return nil, fmt.Errorf("checkout verification timed out after 10 seconds (last status: %s)", lastStatus)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Don't retry 404 (session not found)
	if strings.Contains(errStr, "checkout session not found") || strings.Contains(errStr, "404") {
		return false
	}

	// Don't retry 4xx client errors (except 429)
	if strings.Contains(errStr, "returned status 400") ||
		strings.Contains(errStr, "returned status 401") ||
		strings.Contains(errStr, "returned status 403") {
		return false
	}

	// Retry on:
	// - Network errors
	// - 5xx server errors
	// - 429 rate limit errors
	// - Timeout errors
	// - Connection errors
	return true
}

// IngestMeterEvent ingests a meter event to Polar for usage-based billing
// This notifies Polar about invoice processing to consume meter credits
// Meter: "Invoice Processing"
func (p *polarAdapter) IngestMeterEvent(ctx context.Context, externalCustomerID string, meterSlug string, amount int32) error {
	// Call Polar API to ingest meter event
	// POST /v1/events/ingest endpoint for event ingestion
	endpoint := "/v1/events/ingest"

	// Prepare request body for event ingestion
	// Events must be wrapped in "events" array
	// Meter will automatically aggregate events and decrement credits
	body := map[string]any{
		"events": []map[string]any{
			{
				"name":                 meterSlug,
				"external_customer_id": externalCustomerID,
				"metadata": map[string]any{
					"count": amount,
				},
			},
		},
	}

	// Log the exact payload being sent to Polar for debugging
	bodyJSON, _ := json.Marshal(body)
	fmt.Printf("📤 SENDING TO POLAR - POST %s\nPayload: %s\n", endpoint, string(bodyJSON))

	resp, err := p.client.Post(ctx, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to call Polar events API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("polar events API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Console log for successful event ingestion
	fmt.Printf("✅ EVENT INGESTED - Customer: %s | Event: %s | Amount: %d | Polar meters will aggregate\n",
		externalCustomerID, meterSlug, amount)

	return nil
}

func parseTime(s string) (time.Time, error) {
	// Parse ISO 8601 timestamp
	return time.Parse(time.RFC3339, s)
}
