package cmd

import (
	"fmt"
	"strings"

	"go.uber.org/dig"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	"github.com/moasq/go-b2b-starter/internal/files/config"
	"github.com/moasq/go-b2b-starter/internal/files/domain"
	"github.com/moasq/go-b2b-starter/internal/files/internal/infra"
	"github.com/moasq/go-b2b-starter/internal/logger"
)

func SetupDependencies(container *dig.Container) error {
	// Provider for R2 repository with development mode support
	if err := container.Provide(func(cfg *config.Config, log logger.Logger) (domain.R2Repository, error) {
		// Check for placeholder credentials (development mode)
		if isPlaceholderR2Credentials(cfg) {
			log.Warn("R2 credentials are placeholders - using mock file storage (development mode)", map[string]any{
				"account_id": cfg.R2.AccountID,
				"message":    "File upload/download will not work. Update R2_* variables in app.env with real credentials",
			})
			// Return mock repository for development mode
			return infra.NewMockR2Repository(log), nil
		}

		return infra.NewR2Repository(cfg)
	}); err != nil {
		fmt.Printf("Error providing R2 repository: %v", err)
		return err
	}

	// Provider for database metadata repository
	if err := container.Provide(func(store adapters.FileAssetStore) domain.FileMetadataRepository {
		return infra.NewDBRepository(store)
	}); err != nil {
		fmt.Printf("Error providing database repository: %v", err)
		return err
	}

	// Provider for composite file repository
	if err := container.Provide(infra.NewCompositeRepository); err != nil {
		fmt.Printf("Error providing composite file repository: %v", err)
		return err
	}

	// Provider for file service
	if err := container.Provide(domain.NewFileService); err != nil {
		fmt.Printf("Error providing file service: %v", err)
		return err
	}

	return nil
}

// isPlaceholderR2Credentials checks if the R2 credentials are placeholder values.
func isPlaceholderR2Credentials(cfg *config.Config) bool {
	return strings.Contains(cfg.R2.AccountID, "REPLACE") ||
		strings.Contains(cfg.R2.AccessKeyID, "REPLACE") ||
		strings.Contains(cfg.R2.SecretAccessKey, "REPLACE") ||
		cfg.R2.AccountID == "" ||
		cfg.R2.AccessKeyID == ""
}
