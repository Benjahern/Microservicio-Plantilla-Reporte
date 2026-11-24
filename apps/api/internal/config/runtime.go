package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr, Environment, Version       string
	DatabaseURL                          string
	DatabaseTimeout                      time.Duration
	JWKSURL, Issuer, Audience, RoleClaim string
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr: value("HTTP_ADDR", ":8000"), Environment: value("API_ENV", "dev"), Version: value("APP_VERSION", "dev"),
		DatabaseURL: value("DATABASE_URL", ""), JWKSURL: value("AUTH_JWKS_URL", ""),
		Issuer: value("AUTH_ISSUER", ""), Audience: value("AUTH_AUDIENCE", ""), RoleClaim: value("AUTH_ROLE_CLAIM", "perfil"),
	}
	_, port, err := net.SplitHostPort(c.HTTPAddr)
	n, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || n < 1 || n > 65535 {
		return c, fmt.Errorf("HTTP_ADDR debe tener formato host:puerto")
	}
	if c.Environment != "dev" && c.Environment != "prod" {
		return c, fmt.Errorf("API_ENV debe ser dev o prod")
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL es obligatoria")
	}
	c.DatabaseTimeout, err = time.ParseDuration(value("DATABASE_TIMEOUT", "5s"))
	if err != nil || c.DatabaseTimeout <= 0 || c.DatabaseTimeout > 30*time.Second {
		return c, fmt.Errorf("DATABASE_TIMEOUT debe ser mayor que cero y hasta 30s")
	}
	keyURL, err := url.Parse(c.JWKSURL)
	if err != nil || keyURL.Hostname() == "" || keyURL.User != nil || keyURL.Fragment != "" ||
		(keyURL.Scheme != "https" && (c.Environment != "dev" || keyURL.Scheme != "http")) {
		return c, fmt.Errorf("AUTH_JWKS_URL debe ser HTTPS (HTTP permitido sólo en dev)")
	}
	if c.Issuer == "" || c.Audience == "" {
		return c, fmt.Errorf("AUTH_ISSUER y AUTH_AUDIENCE son obligatorias")
	}
	return c, nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
