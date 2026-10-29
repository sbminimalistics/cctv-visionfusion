package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type RabbitMQConfig struct {
	Host     string `yaml:"host" env:"RMQ_HOST" env-default:"localhost"`
	Port     port   `yaml:"port" env:"RMQ_PORT" env-default:"5672"`
	User     string `yaml:"user" env:"DOWNLOADER_RMQ_USER" env-default:"downloader_worker"`
	Password string `yaml:"password" env:"DOWNLOADER_RMQ_PASS"`
	VHost    string `yaml:"vhost" env:"RMQ_VHOST" env-default:"/cctv-vf"`
}

type port int

type AppConfig struct {
	RabbitMQ       RabbitMQConfig `yaml:"rabbitmq"`
	RabbitMQQueue  string         `env:"RMQ_QUEUE_DOWNLOADER_FEED"`
	BackendBaseURL string         `env:"DOWNLOADER_BACKEND_BASE_URL"`
}

var Config AppConfig

func init() {
	var config AppConfig

	err := cleanenv.ReadConfig("config.yml", &config)
	if err != nil {
		fmt.Println("config load error:", err)
	}

	_ = cleanenv.ReadConfig(".env", &config)

	// This forces cleanenv to re-parse the active environment memory strings
	// and lock down the final production/dev overrides into the final structure.
	err = cleanenv.ReadEnv(&config)
	if err != nil {
		fmt.Println("Environment variable sync error:", err)
	} else {
		fmt.Println("Configuration successfully synchronized across all layers!")
	}

	Config = config
}
