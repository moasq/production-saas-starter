package api

import (
	"github.com/moasq/go-b2b-starter/internal/modules/billing"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations"
	server "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
	"go.uber.org/dig"
)

func Init(container *dig.Container) error {
	if err := billing.RegisterHandlers(container); err != nil {
		return err
	}
	return container.Invoke(func(srv server.Server, org *organizations.Routes, bill *billing.Handler) {
		srv.RegisterRoutes(org.Routes, server.ApiPrefix)
		srv.RegisterRoutes(bill.Routes, server.ApiPrefix)
	})
}
