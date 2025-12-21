package files

import (
	"strings"
)

// File Type Categories
type FileCategory string

const (
	CategoryDocument FileCategory = "document"
	CategoryImage    FileCategory = "image"
	CategoryArchive  FileCategory = "archive"
)

// Supported file types
// SECURITY: Restricted to invoice-safe formats only (PDF and common image formats)
// Removed: Office documents (.doc, .docx, .xls, .xlsx), text files (.txt, .csv),
//          archives (.zip, .rar, etc.), and risky image formats (.svg, .gif)
var (
	DocumentTypes = []string{".pdf"}
	ImageTypes    = []string{".jpg", ".jpeg", ".png"}
	ArchiveTypes  = []string{} // Archives disabled for security
)

// Business file contexts
type FileContext string

const (
	ContextInvoice            FileContext = "invoice"
	ContextReceipt            FileContext = "receipt"
	ContextContract           FileContext = "contract"
	ContextReport             FileContext = "report"
	ContextProfile            FileContext = "profile"
	ContextGeneral            FileContext = "general"
	ContextPaymentInstruction FileContext = "payment_instruction"
	ContextPaymentBatch       FileContext = "payment_batch"
)

// File size limits (in bytes)
// SECURITY: Strict limits for invoice processing to minimize attack surface
const (
	MaxDocumentSize = 2 * 1024 * 1024 // 2MB - sufficient for most invoice PDFs
	MaxImageSize    = 1 * 1024 * 1024 // 1MB - sufficient for scanned invoices
	MaxArchiveSize  = 0               // Archives disabled
)

// GetFileCategory determines the category based on file extension
func GetFileCategory(filename string) FileCategory {
	ext := strings.ToLower(getFileExtension(filename))
	
	for _, docType := range DocumentTypes {
		if ext == docType {
			return CategoryDocument
		}
	}
	
	for _, imgType := range ImageTypes {
		if ext == imgType {
			return CategoryImage
		}
	}
	
	for _, archType := range ArchiveTypes {
		if ext == archType {
			return CategoryArchive
		}
	}
	
	return CategoryDocument // Default to document
}

// GetMaxFileSize returns the maximum file size for a given category
func GetMaxFileSize(category FileCategory) int64 {
	switch category {
	case CategoryImage:
		return MaxImageSize
	case CategoryArchive:
		return MaxArchiveSize
	default:
		return MaxDocumentSize
	}
}

// IsAllowedFileType checks if the file type is allowed
func IsAllowedFileType(filename string) bool {
	ext := strings.ToLower(getFileExtension(filename))
	
	allTypes := append(DocumentTypes, ImageTypes...)
	allTypes = append(allTypes, ArchiveTypes...)
	
	for _, allowedType := range allTypes {
		if ext == allowedType {
			return true
		}
	}
	
	return false
}

func getFileExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[idx:]
	}
	return ""
}