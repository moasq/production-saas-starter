package bootstrap

import (
	"github.com/moasq/go-b2b-starter/internal/api"
	db "github.com/moasq/go-b2b-starter/internal/db/cmd"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	billing "github.com/moasq/go-b2b-starter/internal/modules/billing/cmd"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations"
	"github.com/moasq/go-b2b-starter/internal/platform/betterauth"
	logger "github.com/moasq/go-b2b-starter/internal/platform/logger/cmd"
	polar "github.com/moasq/go-b2b-starter/internal/platform/polar/cmd"
	server "github.com/moasq/go-b2b-starter/internal/platform/server/cmd"
	"go.uber.org/dig"
)

func InitMods(container *dig.Container) {
	server.Init(container)
	logger.Init(container)
	must(db.ProvideDependencies(container))
	must(polar.Init(container))
	for _, provider := range []any{betterauth.NewFromEnvironment, func(c *betterauth.Client) auth.AuthProvider { return c }, organizations.NewResolver, auth.NewMiddleware, organizations.NewRoutes} {
		must(container.Provide(provider))
	}
	must(auth.RegisterNamedMiddlewares(container))
	must(billing.Init(container))
	must(api.Init(container))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
