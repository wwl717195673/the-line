package main

import (
	"log"

	"the-line/backend/internal/app"
	"the-line/backend/internal/config"
	"the-line/backend/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}

	if cfg.AutoMigrate {
		if err := db.AutoMigrate(database); err != nil {
			log.Fatalf("auto migrate: %v", err)
		}
	}

	server := app.NewServer(cfg, database)
	if err := server.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
