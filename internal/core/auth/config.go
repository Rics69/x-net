package core_auth

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	JWTSecret string        `envconfig:"JWT_SECRET" required:"true"`
	TokenTTL  time.Duration `envconfig:"TOKEN_TTL" default:"24h"`

	// Secure-кука уходит только по HTTPS. Локально у нас http://localhost, поэтому false,
	// на проде за HTTPS обязательно true
	CookieSecure bool `envconfig:"COOKIE_SECURE" default:"false"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("AUTH", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Auth config: %w", err)
		panic(err)
	}

	return config
}
