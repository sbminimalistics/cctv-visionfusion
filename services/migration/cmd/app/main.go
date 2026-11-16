package main

import (
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Database driver for migrations
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"sbminimalistics.com/cctv-visionfusion/migration/internal/config"
	"sbminimalistics.com/cctv-visionfusion/migration/sql"
)

func runSchemaMigration(schema string, migrationPath string, dsn string) {
	d, err := iofs.New(sql.Files, migrationPath)
	if err != nil {
		log.Fatalf("failed to load embedded files: %v", err)
	}

	// Append search_path parameter cleanly to target the explicit schema
	targetDSN := fmt.Sprintf("%s&search_path=%s", dsn, schema)

	m, err := migrate.NewWithSourceInstance("iofs", d, targetDSN)
	if err != nil {
		log.Fatalf("migration init failed for %s: %v", schema, err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed executing migration for %s: %v", schema, err)
	}

	fmt.Printf("Successfully migrated schema: %s\n", schema)
}

func main() {
	baseDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Config.DB.User,
		config.Config.DB.Password,
		config.Config.DB.Host,
		config.Config.DB.Port,
		config.Config.DB.Name,
	)

	// orchestrate schemas
	runSchemaMigration(config.Config.DB.BackendSchema, "backend", baseDSN)
	runSchemaMigration(config.Config.DB.DashboardSchema, "dashboard", baseDSN)
}
