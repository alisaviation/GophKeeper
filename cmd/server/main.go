package server

import (
	"log"

	"github.com/alisaviation/GophKeeper/internal/server/storage"
	"github.com/alisaviation/GophKeeper/internal/server/storage/postgres"
)

func main() {
	dbConfig := postgres.DefaultConfig()

	if err := postgres.RunMigrations(dbConfig.DSN()); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	storage, err := storage.NewStorage(storage.Config{
		Type:     storage.TypePostgreSQL,
		Postgres: dbConfig,
	})
	if err != nil {
		log.Fatal("Failed to create storage:", err)
	}
	defer storage.Close()

}
