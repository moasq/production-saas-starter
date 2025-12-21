package eventbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Common event types used across modules

// Invoice Events
type InvoiceUploaded struct {
	BaseEvent
	InvoiceID int32  `json:"invoice_id"`
	FileID    int32  `json:"file_id"`
	VendorName string `json:"vendor_name,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	UserID     int32  `json:"user_id"`
}

func NewInvoiceUploaded(invoiceID, fileID, userID int32, vendorName string, amount decimal.Decimal) *InvoiceUploaded {
	return &InvoiceUploaded{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "invoice.uploaded",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:  invoiceID,
		FileID:     fileID,
		VendorName: vendorName,
		Amount:     amount,
		UserID:     userID,
	}
}

type InvoiceValidated struct {
	BaseEvent
	InvoiceID     int32                  `json:"invoice_id"`
	FileID        int32                  `json:"file_id"`
	ValidationData map[string]interface{} `json:"validation_data"`
}

func NewInvoiceValidated(invoiceID, fileID int32, validationData map[string]interface{}) *InvoiceValidated {
	return &InvoiceValidated{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "invoice.validated",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:      invoiceID,
		FileID:         fileID,
		ValidationData: validationData,
	}
}

// OCR Events
type OCRRequested struct {
	BaseEvent
	InvoiceID int32 `json:"invoice_id"`
	FileID    int32 `json:"file_id"`
}

func NewOCRRequested(invoiceID, fileID int32) *OCRRequested {
	return &OCRRequested{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "ocr.requested",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID: invoiceID,
		FileID:    fileID,
	}
}

type TextExtracted struct {
	BaseEvent
	InvoiceID     int32                  `json:"invoice_id"`
	FileID        int32                  `json:"file_id"`
	ExtractedData map[string]interface{} `json:"extracted_data"`
	Confidence    float64                `json:"confidence"`
}

func NewTextExtracted(invoiceID, fileID int32, extractedData map[string]interface{}, confidence float64) *TextExtracted {
	return &TextExtracted{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "text.extracted",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:     invoiceID,
		FileID:        fileID,
		ExtractedData: extractedData,
		Confidence:    confidence,
	}
}

// Duplicate Detection Events
type DuplicateCheckRequested struct {
	BaseEvent
	InvoiceID int32                  `json:"invoice_id"`
	Data      map[string]interface{} `json:"data"`
}

func NewDuplicateCheckRequested(invoiceID int32, data map[string]interface{}) *DuplicateCheckRequested {
	return &DuplicateCheckRequested{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "duplicate.check_requested",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID: invoiceID,
		Data:      data,
	}
}

type DuplicateDetected struct {
	BaseEvent
	InvoiceID         int32   `json:"invoice_id"`
	DuplicateOf       int32   `json:"duplicate_of"`
	SimilarityScore   float64 `json:"similarity_score"`
	RequiresReview    bool    `json:"requires_review"`
}

func NewDuplicateDetected(invoiceID, duplicateOf int32, similarityScore float64, requiresReview bool) *DuplicateDetected {
	return &DuplicateDetected{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "duplicate.detected",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:       invoiceID,
		DuplicateOf:     duplicateOf,
		SimilarityScore: similarityScore,
		RequiresReview:  requiresReview,
	}
}

type UniqueConfirmed struct {
	BaseEvent
	InvoiceID int32 `json:"invoice_id"`
}

func NewUniqueConfirmed(invoiceID int32) *UniqueConfirmed {
	return &UniqueConfirmed{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "duplicate.unique_confirmed",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID: invoiceID,
	}
}

// Approval Events
type ApprovalRequested struct {
	BaseEvent
	InvoiceID      int32           `json:"invoice_id"`
	Amount         decimal.Decimal `json:"amount"`
	VendorID       int32           `json:"vendor_id"`
	RequesterID    int32           `json:"requester_id"`
	ApprovalLevel  int             `json:"approval_level"`
}

func NewApprovalRequested(invoiceID, vendorID, requesterID int32, amount decimal.Decimal, approvalLevel int) *ApprovalRequested {
	return &ApprovalRequested{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "approval.requested",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:     invoiceID,
		Amount:        amount,
		VendorID:      vendorID,
		RequesterID:   requesterID,
		ApprovalLevel: approvalLevel,
	}
}

type ApprovalGranted struct {
	BaseEvent
	InvoiceID   int32  `json:"invoice_id"`
	ApproverID  int32  `json:"approver_id"`
	ApprovalID  int32  `json:"approval_id"`
	Comments    string `json:"comments,omitempty"`
}

func NewApprovalGranted(invoiceID, approverID, approvalID int32, comments string) *ApprovalGranted {
	return &ApprovalGranted{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "approval.granted",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:  invoiceID,
		ApproverID: approverID,
		ApprovalID: approvalID,
		Comments:   comments,
	}
}

type ApprovalRejected struct {
	BaseEvent
	InvoiceID   int32  `json:"invoice_id"`
	ApproverID  int32  `json:"approver_id"`
	ApprovalID  int32  `json:"approval_id"`
	Reason      string `json:"reason"`
}

func NewApprovalRejected(invoiceID, approverID, approvalID int32, reason string) *ApprovalRejected {
	return &ApprovalRejected{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "approval.rejected",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:  invoiceID,
		ApproverID: approverID,
		ApprovalID: approvalID,
		Reason:     reason,
	}
}

// Payment Events
type PaymentScheduled struct {
	BaseEvent
	InvoiceID         int32           `json:"invoice_id"`
	PaymentID         int32           `json:"payment_id"`
	ScheduledDate     time.Time       `json:"scheduled_date"`
	Amount            decimal.Decimal `json:"amount"`
	DiscountCaptured  decimal.Decimal `json:"discount_captured"`
	OptimalPayment    bool            `json:"optimal_payment"`
}

func NewPaymentScheduled(invoiceID, paymentID int32, scheduledDate time.Time, amount, discountCaptured decimal.Decimal, optimalPayment bool) *PaymentScheduled {
	return &PaymentScheduled{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "payment.scheduled",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:        invoiceID,
		PaymentID:        paymentID,
		ScheduledDate:    scheduledDate,
		Amount:           amount,
		DiscountCaptured: discountCaptured,
		OptimalPayment:   optimalPayment,
	}
}

type PaymentExecuted struct {
	BaseEvent
	InvoiceID         int32           `json:"invoice_id"`
	PaymentID         int32           `json:"payment_id"`
	OrganizationID    int32           `json:"organization_id"`
	TransactionID     string          `json:"transaction_id"`
	Amount            decimal.Decimal `json:"amount"`
	DiscountCaptured  decimal.Decimal `json:"discount_captured"`
	ExecutedDate      time.Time       `json:"executed_date"`
}

func NewPaymentExecuted(invoiceID, paymentID, organizationID int32, transactionID string, amount, discountCaptured decimal.Decimal, executedDate time.Time) *PaymentExecuted {
	return &PaymentExecuted{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Name:      "payment.executed",
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		InvoiceID:        invoiceID,
		PaymentID:        paymentID,
		OrganizationID:   organizationID,
		TransactionID:    transactionID,
		Amount:           amount,
		DiscountCaptured: discountCaptured,
		ExecutedDate:     executedDate,
	}
}