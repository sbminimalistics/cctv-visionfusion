package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC search_path=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("MIGRATION_DB_USER", "usr"),
		getEnv("MIGRATION_DB_PASS", "pass"),
		getEnv("DB_NAME", "cctv"),
		getEnv("DB_PORT", "5432"),
		getEnv("BACKEND_DB_SCHEMA", "db_schema"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database for codegen: %v", err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:           "internal/dal/query",
		ModelPkgPath:      "internal/dal/model",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true, // Generates *string / *time.Time for nullable columns (like trace_id!)
		FieldWithIndexTag: true, // Preserves index tags on struct fields
		FieldWithTypeTag:  true, // Preserves exact SQL column types
	})

	g.UseDB(db)

	// fetch all tables in the public schema except the migration bookkeeping table
	tables, err := db.Migrator().GetTables()
	if err != nil {
		log.Fatalf("failed to list tables: %v", err)
	}

	for _, table := range tables {
		if table == "schema_migrations" {
			continue
		}
		g.GenerateModel(table)
	}

	// Execute code generation
	g.Execute()
}
