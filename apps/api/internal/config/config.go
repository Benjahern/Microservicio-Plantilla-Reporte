package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Identity struct {
	Token   string `json:"token"`
	UserID  string `json:"usuario_id"`
	Profile string `json:"perfil"`
}

type Config struct {
	HTTPAddr, MongoURI, MongoDatabase string
	MongoTimeout                      time.Duration
	Identities                        []Identity
}

func Load() (Config, error) {
	c := Config{HTTPAddr: value("HTTP_ADDR", ":8080"), MongoURI: value("MONGO_URI", "mongodb://localhost:27017/?replicaSet=rs0"), MongoDatabase: value("MONGO_DATABASE", "plantillas_reportes")}
	d, err := time.ParseDuration(value("MONGO_TIMEOUT", "10s"))
	if err != nil || d <= 0 {
		return c, fmt.Errorf("MONGO_TIMEOUT debe ser una duración positiva")
	}
	c.MongoTimeout = d
	raw := os.Getenv("AUTH_IDENTITIES_JSON")
	if raw == "" {
		return c, fmt.Errorf("AUTH_IDENTITIES_JSON es obligatorio")
	}
	if err := json.Unmarshal([]byte(raw), &c.Identities); err != nil || len(c.Identities) == 0 {
		return c, fmt.Errorf("AUTH_IDENTITIES_JSON debe ser un arreglo JSON no vacío válido")
	}
	tokens := map[string]bool{}
	for i := range c.Identities {
		x := &c.Identities[i]
		x.Token = strings.TrimSpace(x.Token)
		x.UserID = strings.TrimSpace(x.UserID)
		if x.Token == "" || x.UserID == "" {
			return c, fmt.Errorf("identidad %d requiere token y usuario_id", i)
		}
		if x.Profile != "lector" && x.Profile != "editor" && x.Profile != "administrador" {
			return c, fmt.Errorf("perfil inválido en identidad %d", i)
		}
		if tokens[x.Token] {
			return c, fmt.Errorf("tokens duplicados en AUTH_IDENTITIES_JSON")
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
