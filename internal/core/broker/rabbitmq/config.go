package core_rabbitmq

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// выключено - события живут только внутри одного инстанса (как было до брокера).
	// Удобно для локальной разработки без лишнего контейнера
	Enabled bool `envconfig:"ENABLED" default:"false"`

	Host     string `envconfig:"HOST" default:"localhost"`
	Port     int    `envconfig:"PORT" default:"5672"`
	User     string `envconfig:"USER" default:"guest"`
	Password string `envconfig:"PASSWORD" default:"guest"`
	VHost    string `envconfig:"VHOST" default:"/"`

	ReconnectMaxDelay time.Duration `envconfig:"RECONNECT_MAX_DELAY" default:"30s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("RABBITMQ", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get RabbitMQ config: %w", err)
		panic(err)
	}

	return config
}
