# File Manager Guide

Simple guide for uploading and managing files with Cloudflare R2 (S3-compatible storage).

## Setup

### R2 Configuration

Add to your `.env`:

```bash
R2_ACCOUNT_ID=your-cloudflare-account-id
R2_ACCESS_KEY_ID=your-r2-access-key
R2_SECRET_ACCESS_KEY=your-r2-secret-key
R2_BUCKET=your-bucket-name
R2_REGION=auto                    # Default
```

### Get R2 Credentials

1. Go to Cloudflare Dashboard → R2
2. Create a bucket or use existing one
3. Go to "Manage R2 API Tokens"
4. Create API token with read/write permissions
5. Copy Account ID, Access Key ID, and Secret Access Key

## Usage in Your Module

### 1. Inject File Service

```go
import "github.com/moasq/go-b2b-starter/pkg/file_manager/domain"

type InvoiceService struct {
    fileService domain.FileService
}

func NewInvoiceService(fileService domain.FileService) *InvoiceService {
    return &InvoiceService{fileService: fileService}
}
```

### 2. Upload a File

```go
func (s *InvoiceService) UploadInvoice(ctx context.Context, file io.Reader, filename string, size int64) (*domain.FileAsset, error) {
    // Create upload request
    req := &domain.FileUploadRequest{
        Filename:    filename,
        Size:        size,
        ContentType: "application/pdf",
        Context:     file_manager.ContextInvoice, // Business context
        Metadata: map[string]any{
            "uploaded_by": userID,
            "invoice_number": "INV-2024-001",
        },
    }

    // Upload to R2
    fileAsset, err := s.fileService.UploadFile(ctx, req, file)
    if err != nil {
        return nil, fmt.Errorf("upload failed: %w", err)
    }

    s.logger.Info("File uploaded", map[string]any{
        "file_id": fileAsset.ID,
        "size":    fileAsset.Size,
        "path":    fileAsset.StoragePath,
    })

    return fileAsset, nil
}
```

### 3. Download a File

```go
func (s *InvoiceService) DownloadInvoice(ctx context.Context, fileID int32) (io.ReadCloser, error) {
    content, fileAsset, err := s.fileService.DownloadFile(ctx, fileID)
    if err != nil {
        return nil, fmt.Errorf("download failed: %w", err)
    }
    defer content.Close()

    // Use the file content
    data, err := io.ReadAll(content)
    if err != nil {
        return nil, err
    }

    return content, nil
}
```

### 4. Get Presigned URL

Get a temporary signed URL for direct browser access:

```go
func (s *InvoiceService) GetInvoiceURL(ctx context.Context, fileID int32) (string, error) {
    // Generate URL valid for 24 hours
    url, err := s.fileService.GetFileURL(ctx, fileID, 24)
    if err != nil {
        return "", err
    }

    return url, nil
}
```

### 5. Delete a File

```go
func (s *InvoiceService) DeleteInvoice(ctx context.Context, fileID int32) error {
    return s.fileService.DeleteFile(ctx, fileID)
}
```

### 6. List Files with Filter

```go
func (s *InvoiceService) ListInvoices(ctx context.Context) ([]*domain.FileAsset, error) {
    // Filter by context
    invoiceContext := file_manager.ContextInvoice

    filter := &domain.FileSearchFilter{
        Context: &invoiceContext,
    }

    files, err := s.fileService.ListFiles(ctx, filter, 50, 0)
    if err != nil {
        return nil, err
    }

    return files, nil
}
```

## File Contexts

Organize files by business purpose:

```go
file_manager.ContextInvoice              // Invoices
file_manager.ContextReceipt              // Receipts
file_manager.ContextContract             // Contracts
file_manager.ContextReport               // Reports
file_manager.ContextProfile              // User profiles
file_manager.ContextPaymentInstruction   // Payment instructions
file_manager.ContextPaymentBatch         // Payment batches
file_manager.ContextGeneral              // General files
```

## File Categories & Limits

**Documents (PDFs):**
- Allowed: `.pdf`
- Max size: 2 MB
- Category: `file_manager.CategoryDocument`

**Images:**
- Allowed: `.jpg`, `.jpeg`, `.png`
- Max size: 1 MB
- Category: `file_manager.CategoryImage`

## Security Features

The file manager automatically:
- ✅ Sanitizes filenames (prevents path traversal)
- ✅ Validates file types (magic byte detection)
- ✅ Enforces size limits
- ✅ Validates content matches extension
- ✅ Stores metadata in PostgreSQL
- ✅ Organizes files in R2 by context and date

## Real-World Example: Complete Upload Flow

```go
func (s *InvoiceService) ProcessInvoiceUpload(ctx context.Context, r *http.Request) (*Invoice, error) {
    // 1. Parse multipart form
    file, header, err := r.FormFile("invoice")
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // 2. Upload to R2 via file manager
    req := &domain.FileUploadRequest{
        Filename:    header.Filename,
        Size:        header.Size,
        ContentType: header.Header.Get("Content-Type"),
        Context:     file_manager.ContextInvoice,
        Metadata: map[string]any{
            "uploaded_by": ctx.Value("user_id"),
            "organization_id": ctx.Value("organization_id"),
        },
    }

    fileAsset, err := s.fileService.UploadFile(ctx, req, file)
    if err != nil {
        return nil, err
    }

    // 3. Create invoice record with file reference
    invoice := &Invoice{
        Number:         "INV-2024-001",
        FileID:         fileAsset.ID,
        FileName:       fileAsset.Filename,
        OrganizationID: organizationID,
    }

    err = s.repo.CreateInvoice(ctx, invoice)
    if err != nil {
        // Rollback: delete the uploaded file
        s.fileService.DeleteFile(ctx, fileAsset.ID)
        return nil, err
    }

    return invoice, nil
}
```

## Storage Structure

Files are organized in R2 with this pattern:
```
{category}/{context}/{date}/{filename}
```

Example:
```
document/invoice/2024/12/09/invoice-12345.pdf
image/receipt/2024/12/09/receipt-photo.jpg
```

## Configuration Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `R2_ACCOUNT_ID` | Yes | Your Cloudflare account ID |
| `R2_ACCESS_KEY_ID` | Yes | R2 API access key ID |
| `R2_SECRET_ACCESS_KEY` | Yes | R2 API secret key |
| `R2_BUCKET` | Yes | R2 bucket name |
| `R2_REGION` | No | Region (default: `auto`) |

## Best Practices

**1. Always use contexts:**
```go
// ✅ Good - organized by purpose
Context: file_manager.ContextInvoice

// ❌ Bad - generic context
Context: file_manager.ContextGeneral
```

**2. Store file references:**
```go
type Invoice struct {
    FileID   int32  `json:"file_id"`
    FileName string `json:"file_name"`
}
```

**3. Handle deletions:**
```go
// Delete invoice and its file
s.invoiceRepo.Delete(ctx, invoiceID)
s.fileService.DeleteFile(ctx, invoice.FileID)
```

**4. Use presigned URLs for downloads:**
```go
// Generate temporary URL instead of downloading in backend
url, _ := s.fileService.GetFileURL(ctx, fileID, 1) // 1 hour
// Return URL to frontend
```

## Why R2?

- **S3-compatible**: Use standard AWS SDK
- **No egress fees**: Free bandwidth
- **Global CDN**: Fast downloads worldwide
- **Cost-effective**: Lower storage costs than S3

That's it! Just inject `FileService` and manage files with R2.
