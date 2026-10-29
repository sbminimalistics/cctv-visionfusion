package config

import (
	"fmt"
	"regexp"

	"github.com/ilyakaznacheev/cleanenv"
)

type DeviceType string

const (
	Camera DeviceType = "camera"
	NVR    DeviceType = "nvr"
)

type Manufacturer string

const (
	Hikvision Manufacturer = "hikvision"
)

type EdgeDeviceConfig struct {
	Id                  string       `yaml:"id"`
	Protocol            string       `yaml:"protocol"`
	Host                string       `yaml:"host"`
	Port                port         `yaml:"port"`
	User                string       `yaml:"user"`
	Password            string       `yaml:"password"`
	Manufacturer        Manufacturer `yaml:"manufacturer"`
	DeviceType          DeviceType   `yaml:"device_type"`
	DestinationExchange string       `yaml:"destination_exchange"`
	PicturePath         string       `yaml:"picture_path"`
}

type CompositionSourceConfig struct {
	DeviceID  string `yaml:"device_id"`
	EventType string `yaml:"event_type"`
}

type EventCompositionConfig struct {
	Id                      string                    `yaml:"id"`
	DefaultEventRetentionMS int32                     `yaml:"default_event_retention_ms"`
	Sources                 []CompositionSourceConfig `yaml:"sources"`
}

type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     port   `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name     string `yaml:"name" env:"DB_NAME" env-default:"postgres"`
	User     string `yaml:"user" env:"BACKEND_DB_USER" env-default:"user"`
	Password string `yaml:"password" env:"BACKEND_DB_PASS"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host" env:"RMQ_HOST" env-default:"localhost"`
	Port     port   `yaml:"port" env:"RMQ_PORT" env-default:"5672"`
	User     string `yaml:"user" env:"BACKEND_RMQ_USER" env-default:"backend_worker"`
	Password string `yaml:"password" env:"BACKEND_RMQ_PASS"`
	VHost    string `yaml:"vhost" env:"RMQ_VHOST" env-default:"/cctv-vf"`
}

type RouteSourceEventRegexp struct {
	*regexp.Regexp
}

func (r *RouteSourceEventRegexp) UnmarshalText(text []byte) error {
	re, err := regexp.Compile(string(text))
	if err != nil {
		return fmt.Errorf("invalid regular expression in yaml: %w", err)
	}
	r.Regexp = re
	return nil
}

func (r *RouteSourceEventRegexp) SetValue(s string) error {
	re, err := regexp.Compile(s)
	if err != nil {
		return fmt.Errorf("invalid regular expression in env: %w", err)
	}
	r.Regexp = re
	return nil
}

type RouteSource struct {
	SourceID        string                  `yaml:"source_id"`
	EventTypeRegexp *RouteSourceEventRegexp `yaml:"event_type_regexp"`
}

type Route struct {
	Source              RouteSource `yaml:"source"`
	DestinationExchange string      `yaml:"destination_exchange"`
}

type port int

type APIConfig struct {
	Port port `yaml:"port" env:"BACKEND_API_PORT"`
}

type AppConfig struct {
	EdgeDevices       []EdgeDeviceConfig       `yaml:"edge_devices"`
	EventCompositions []EventCompositionConfig `yaml:"event_compositions"`
	Routing           []Route                  `yaml:"routing"`
	DB                DBConfig                 `yaml:"database"`
	RabbitMQ          RabbitMQConfig           `yaml:"rabbitmq"`
	API               APIConfig                `yaml:"api"`
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
