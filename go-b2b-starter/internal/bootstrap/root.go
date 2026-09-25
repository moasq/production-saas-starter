package bootstrap

import (
	"github.com/moasq/go-b2b-starter/internal/db/postgres"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/dig"

	server "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

func Execute() {
	if err := godotenv.Load("app.env"); err != nil {
		log.Printf("Warning: Error loading app.env file: %v", err)
	}

	if len(os.Args) > 1 {
		if len(os.Args) != 2 || os.Args[1] != "migrate" {
			log.Fatal("usage: api [migrate]")
		}
		cfg, err := postgres.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		pool, err := postgres.InitDB(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()
		if err := postgres.NewPostgresManager(cfg, pool).RunMigrations(); err != nil {
			log.Fatal(err)
		}
		log.Print("database migrations complete")
		return
	}
	container := dig.New()

	InitMods(container)

	var srv server.Server

	if err := container.Invoke(func(s server.Server) {
		srv = s
	}); err != nil {
		panic(err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
