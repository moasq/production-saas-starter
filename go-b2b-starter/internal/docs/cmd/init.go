package cmd

import (
	"log"

	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/docs/api"
	server "github.com/moasq/go-b2b-starter/internal/server/domain"
)

func Init(container *dig.Container) {
	err := container.Invoke(func(srv server.Server) {
		handler := api.NewHandler()
		srv.RegisterRoutes(handler.Routes, "")
	})

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
