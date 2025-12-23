package bootstrap

import (
	"log"

	"github.com/joho/godotenv"
	"go.uber.org/dig"

	server "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

func Execute() {
	if err := godotenv.Load("app.env"); err != nil {
		log.Printf("Warning: Error loading app.env file: %v", err)
	}

	container := dig.New()

	InitMods(container)

	var srv server.Server

	if err := container.Invoke(func(s server.Server) {
		srv = s
	}); err != nil {
		panic(err)
	}

	srv.Start()

}
