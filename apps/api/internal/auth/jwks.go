package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/example/microservicio-plantilla-reporte/internal/usuarios"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("token inválido")
var ErrIdentityUnavailable = errors.New("proveedor de identidad no disponible")

type JWT struct {
	keys      *keySet
	parser    *jwt.Parser
	roleClaim string
}

func NewJWT(url, issuer, audience, roleClaim string) *JWT {
	return &JWT{
		keys: &keySet{url: url, client: &http.Client{
			Timeout:       5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}},
		parser: jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(issuer),
			jwt.WithAudience(audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(15*time.Second)),
		roleClaim: roleClaim,
	}
}

func (a *JWT) Authenticate(ctx context.Context, raw string) (usuarios.UsuarioDTO, error) {
	if len(raw) > 8192 {
		return usuarios.UsuarioDTO{}, ErrInvalidToken
	}
	claims := jwt.MapClaims{}
	token, err := a.parser.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" || len(kid) > 128 {
			return nil, ErrInvalidToken
		}
		return a.keys.key(ctx, kid)
	})
	if errors.Is(err, ErrIdentityUnavailable) {
		return usuarios.UsuarioDTO{}, ErrIdentityUnavailable
	}
	if err != nil || !token.Valid {
		return usuarios.UsuarioDTO{}, ErrInvalidToken
	}
	id, err := claims.GetSubject()
	role, _ := claims[a.roleClaim].(string)
	if err != nil || !api.UUIDPattern.MatchString(id) || role == "" {
		return usuarios.UsuarioDTO{}, ErrInvalidToken
	}
	return usuarios.UsuarioDTO{ID: strings.ToLower(id), Perfil: role}, nil
}

type keySet struct {
	mu                 sync.Mutex
	url                string
	client             *http.Client
	keys               map[string]*rsa.PublicKey
	expires, lastFetch time.Time
	fetching           chan struct{}
	lastFailed         bool
}

func (s *keySet) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	for {
		s.mu.Lock()
		now := time.Now()
		fresh := now.Before(s.expires)
		if key := s.keys[kid]; fresh && key != nil {
			s.mu.Unlock()
			return key, nil
		}
		if active := s.fetching; active != nil {
			s.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ErrIdentityUnavailable
			case <-active:
				continue
			}
		}
		if now.Sub(s.lastFetch) < 5*time.Second {
			unavailable := !fresh && s.lastFailed
			s.mu.Unlock()
			if unavailable {
				return nil, ErrIdentityUnavailable
			}
			return nil, ErrInvalidToken
		}
		active := make(chan struct{})
		s.fetching = active
		s.mu.Unlock()

		keys, err := s.fetch(ctx)
		s.mu.Lock()
		s.lastFetch, s.lastFailed = time.Now(), err != nil
		if err == nil {
			s.keys, s.expires = keys, s.lastFetch.Add(5*time.Minute)
		}
		s.fetching = nil
		close(active)
		s.mu.Unlock()
	}
}

type jsonKey struct {
	ID   string `json:"kid"`
	Type string `json:"kty"`
	Use  string `json:"use"`
	Alg  string `json:"alg"`
	N    string `json:"n"`
	E    string `json:"e"`
}

func (s *keySet) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, ErrIdentityUnavailable
	}
	const max = 128 << 10
	data, err := io.ReadAll(io.LimitReader(res.Body, max+1))
	if err != nil || len(data) > max {
		return nil, ErrIdentityUnavailable
	}
	var set struct {
		Keys []jsonKey `json:"keys"`
	}
	if err := json.Unmarshal(data, &set); err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, candidate := range set.Keys {
		key, ok := candidate.publicKey()
		if !ok {
			continue
		}
		if keys[candidate.ID] != nil {
			return nil, ErrIdentityUnavailable
		}
		keys[candidate.ID] = key
	}
	if len(keys) == 0 {
		return nil, ErrIdentityUnavailable
	}
	return keys, nil
}

func (k jsonKey) publicKey() (*rsa.PublicKey, bool) {
	if k.ID == "" || len(k.ID) > 128 || k.Type != "RSA" || (k.Use != "" && k.Use != "sig") || (k.Alg != "" && k.Alg != "RS256") {
		return nil, false
	}
	n, nErr := base64.RawURLEncoding.DecodeString(k.N)
	e, eErr := base64.RawURLEncoding.DecodeString(k.E)
	if nErr != nil || eErr != nil || len(n) > 1024 || len(e) > 4 {
		return nil, false
	}
	mod, exponent := new(big.Int).SetBytes(n), new(big.Int).SetBytes(e).Int64()
	if mod.BitLen() < 2048 || mod.Bit(0) == 0 || exponent < 3 || exponent > 2147483647 || exponent%2 == 0 {
		return nil, false
	}
	return &rsa.PublicKey{N: mod, E: int(exponent)}, true
}
