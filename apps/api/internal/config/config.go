package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
)

type Identity struct {
	Token   string `json:"token"`
	UserID  string `json:"usuario_id"`
	Profile string `json:"perfil"`
}

type Config struct {
	HTTPAddr, Environment string
	Identities            []Identity
}

func Load() (Config, error) {
	c := Config{HTTPAddr: value("HTTP_ADDR", ":8000"), Environment: value("API_ENV", "dev")}
	if _, port, err := net.SplitHostPort(c.HTTPAddr); err != nil || port == "" {
		return c, fmt.Errorf("HTTP_ADDR debe tener formato host:puerto")
	}
	if c.Environment != "dev" && c.Environment != "prod" {
		return c, fmt.Errorf("API_ENV debe ser dev o prod")
	}
	raw := os.Getenv("MOCK_IDENTITIES_JSON")
	if raw == "" {
		raw = `[{"token":"demo-editor","usuario_id":"11111111-1111-4111-8111-111111111111","perfil":"editor"},{"token":"demo-lector","usuario_id":"11111111-1111-4111-8111-111111111111","perfil":"lector"},{"token":"demo-admin","usuario_id":"22222222-2222-4222-8222-222222222222","perfil":"administrador"}]`
	}
	if err := json.Unmarshal([]byte(raw), &c.Identities); err != nil || len(c.Identities) == 0 {
		return c, fmt.Errorf("MOCK_IDENTITIES_JSON debe ser un arreglo JSON no vacío válido")
	}
	tokens := map[string]bool{}
	for i := range c.Identities {
		x := &c.Identities[i]
		x.Token = strings.TrimSpace(x.Token)
		x.UserID = strings.ToLower(strings.TrimSpace(x.UserID))
		if x.Token == "" || !api.UUIDPattern.MatchString(x.UserID) {
			return c, fmt.Errorf("identidad %d requiere token y usuario_id UUID", i)
		}
		if x.Profile != "lector" && x.Profile != "editor" && x.Profile != "administrador" {
			return c, fmt.Errorf("perfil inválido en identidad %d", i)
		}
		if tokens[x.Token] {
			return c, fmt.Errorf("tokens duplicados en MOCK_IDENTITIES_JSON")
		}
		tokens[x.Token] = true
	}
	return c, nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
