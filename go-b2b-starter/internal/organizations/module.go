package organizations

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/organizations/app/services"
	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
	"github.com/moasq/go-b2b-starter/internal/organizations/infra/repositories"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/logger/domain"
	stytchcfg "github.com/moasq/go-b2b-starter/internal/stytch"
)

// Module provides organization module dependencies
type Module struct {
	container *dig.Container
}

func NewModule(container *dig.Container) *Module {
	return &Module{
		container: container,
	}
}

// RegisterDependencies registers all organization module dependencies
func (m *Module) RegisterDependencies() error {
	// Register local database repositories
	if err := m.container.Provide(func(
		accountStore adapters.AccountStore,
		orgStore adapters.OrganizationStore,
	) domain.AccountRepository {
		return repositories.NewAccountRepository(accountStore, orgStore)
	}); err != nil {
		return err
	}

	if err := m.container.Provide(func(
		orgStore adapters.OrganizationStore,
	) domain.OrganizationRepository {
		return repositories.NewOrganizationRepository(orgStore)
	}); err != nil {
		return err
	}

	// Register auth provider repositories (Stytch implementation)
	if err := m.container.Provide(func(
		client *stytchcfg.Client,
		logger loggerDomain.Logger,
		localOrgRepo domain.OrganizationRepository,
	) domain.AuthOrganizationRepository {
		return repositories.NewStytchOrganizationRepository(client, logger, localOrgRepo)
	}); err != nil {
		return err
	}

	if err := m.container.Provide(func(
		client *stytchcfg.Client,
		cfg *stytchcfg.Config,
		logger loggerDomain.Logger,
	) domain.AuthMemberRepository {
		return repositories.NewStytchMemberRepository(client, *cfg, logger)
	}); err != nil {
		return err
	}

	if err := m.container.Provide(func(
		client *stytchcfg.Client,
		logger loggerDomain.Logger,
	) domain.AuthRoleRepository {
		return repositories.NewStytchRoleRepository(client, logger)
	}); err != nil {
		return err
	}

	// Register organization service
	if err := m.container.Provide(func(
		orgRepo domain.OrganizationRepository,
		accountRepo domain.AccountRepository,
	) services.OrganizationService {
		return services.NewOrganizationService(orgRepo, accountRepo)
	}); err != nil {
		return err
	}

	// Register member service (for auth member operations)
	if err := m.container.Provide(func(
		authOrgRepo domain.AuthOrganizationRepository,
		authMemberRepo domain.AuthMemberRepository,
		authRoleRepo domain.AuthRoleRepository,
		localOrgRepo domain.OrganizationRepository,
		localAccountRepo domain.AccountRepository,
		logger loggerDomain.Logger,
	) services.MemberService {
		return services.NewMemberService(
			authOrgRepo,
			authMemberRepo,
			authRoleRepo,
			localOrgRepo,
			localAccountRepo,
			logger,
		)
	}); err != nil {
		return err
	}

	return nil
}
