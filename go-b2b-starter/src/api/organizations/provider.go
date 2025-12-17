package organizations

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/app/organizations/app/services"
	"github.com/moasq/go-b2b-starter/pkg/logger"
)

// Provider provides organization API dependencies
type Provider struct {
	container *dig.Container
}

func NewProvider(container *dig.Container) *Provider {
	return &Provider{
		container: container,
	}
}

// RegisterDependencies registers organization API dependencies
func (p *Provider) RegisterDependencies() error {
	// Register handlers
	if err := p.container.Provide(func(
		orgService services.OrganizationService,
		logger logger.Logger,
	) *OrganizationHandler {
		return NewOrganizationHandler(orgService, logger)
	}); err != nil {
		return err
	}

	if err := p.container.Provide(func(
		orgService services.OrganizationService,
		logger logger.Logger,
	) *AccountHandler {
		return NewAccountHandler(orgService, logger)
	}); err != nil {
		return err
	}

	// Register member handler (for auth/member routes)
	if err := p.container.Provide(func(
		memberService services.MemberService,
		logger logger.Logger,
	) *MemberHandler {
		return NewMemberHandler(memberService, logger)
	}); err != nil {
		return err
	}

	// Register routes
	if err := p.container.Provide(func(
		organizationHandler *OrganizationHandler,
		accountHandler *AccountHandler,
		memberHandler *MemberHandler,
	) *Routes {
		return NewRoutes(organizationHandler, accountHandler, memberHandler)
	}); err != nil {
		return err
	}

	return nil
}
