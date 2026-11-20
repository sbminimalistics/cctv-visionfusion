package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type DBConfig struct {
	Host            string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port            port   `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name            string `yaml:"name" env:"DB_NAME" env-default:"postgres"`
	User            string `yaml:"user" env:"MIGRATION_DB_USER" env-default:"user"`
	Password        string `yaml:"password" env:"MIGRATION_DB_PASS"`
	BackendSchema   string `yaml:"backend_schema" env:"BACKEND_DB_SCHEMA"`
	DashboardSchema string `yaml:"dashboard_schema" env:"DASHBOARD_DB_SCHEMA"`
}

type port int

type AppConfig struct {
	DB DBConfig `yaml:"database"`
}

var Config AppConfig

func init() {
	var config AppConfig

	// err := cleanenv.ReadConfig("../../.env", &config)
	// if err != nil {
	// 	fmt.Println("config env file load error:", err)
	// }

	// err := cleanenv.ReadConfig("config.yml", &config)
	// if err != nil {
	// 	fmt.Println("config load error:", err)
	// }

	// This forces cleanenv to re-parse the active environment memory strings
	// and lock down the final production/dev overrides into the final structure.
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		fmt.Println("Environment variable sync error:", err)
	} else {
		fmt.Println("cleanenv finished parsing")
	}

	Config = config
}
