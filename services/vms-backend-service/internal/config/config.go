package config

import (
	"fmt"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
)

type CameraConfig struct {
	Name     string `yaml:"name"`
	IP       string `yaml:"ip"`
	Port     port   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     port   `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name     string `yaml:"name" env:"DB_NAME" env-default:"postgres"`
	User     string `yaml:"user" env:"BACKEND_DB_USER" env-default:"user"`
	Password string `yaml:"password" env:"BACKEND_DB_PASS"`
}

type RabbitMQConfig struct {
	Host             string `yaml:"host" env:"RMQ_HOST" env-default:"localhost"`
	Port             port   `yaml:"port" env:"RMQ_PORT" env-default:"5672"`
	User             string `yaml:"user" env:"BACKEND_RMQ_USER" env-default:"backend_worker"`
	Password         string `yaml:"password" env:"BACKEND_RMQ_PASS"`
	VHost            string `yaml:"vhost" env:"RMQ_VHOST" env-default:"/vms"`
	ExchangeEvents   string `yaml:"exchange_events" env:"RMQ_EXCHANGE_BACKEND_EVENTS"`
	ExchangeTriggers string `yaml:"exchange_triggers" env:"RMQ_EXCHANGE_BACKEND_TRIGGERS"`
}

type AppConfig struct {
	Cameras  []CameraConfig `yaml:"cameras"`
	DB       DBConfig       `yaml:"database"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
}

type port int

func (r *port) SetValue(s string) error {
	if s == "" {
		return fmt.Errorf("field value can't be empty")
	}

	p, err := strconv.Atoi(s)

	if err != nil {
		return fmt.Errorf("provided port value can not be parsed as integer")
	}

	*r = port(p)

	return nil
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
	// and lock down the final production/dev overrides into your structure.
	err = cleanenv.ReadEnv(&config)
	if err != nil {
		fmt.Println("Environment variable sync error:", err)
	} else {
		fmt.Println("Configuration successfully synchronized across all layers!")
	}

	Config = config
}
