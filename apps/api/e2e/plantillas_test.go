package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const base = "/api/v1/plantillas"
const owner = "11111111-1111-4111-8111-111111111111"
const stranger = "22222222-2222-4222-8222-222222222222"
const audience = "plantillas-e2e"

type response struct {
	status int
	header http.Header
	body   []byte
	err    error
}

type server struct{ url string }

func (s server) request(method, path, token, body string, cookie bool, headers ...http.Header) response {
	req, err := http.NewRequest(method, s.url+path, strings.NewReader(body))
	if err != nil {
		return response{err: err}
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, h := range headers {
		for name, values := range h {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}
	}
	if cookie {
		req.AddCookie(&http.Cookie{Name: "sesion", Value: token})
	} else if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return response{err: err}
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	return response{status: res.StatusCode, header: res.Header, body: b, err: err}
}

func (r response) check(t *testing.T, status int) response {
	t.Helper()
	if r.err != nil || r.status != status {
		t.Fatalf("HTTP %d, esperado %d: %s (%v)", r.status, status, r.body, r.err)
	}
	if !api.UUIDPattern.MatchString(r.header.Get("X-Request-Id")) {
		t.Fatal("falta X-Request-Id válido")
	}
	if status >= 400 && r.header.Get("Content-Type") != "application/problem+json" {
		t.Fatal("error sin Problem Details")
	}
	return r
}

func decode[T any](t *testing.T, r response) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(r.body, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func start(t *testing.T, bin, db, issuer string) (server, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "HTTP_ADDR="+addr, "DATABASE_URL="+db, "DATABASE_TIMEOUT=2s",
		"API_ENV=dev", "AUTH_JWKS_URL="+issuer+"/jwks", "AUTH_ISSUER="+issuer, "AUTH_AUDIENCE="+audience,
		"AUTH_ROLE_CLAIM=perfil", "APP_VERSION=e2e")
	capture := os.Getenv("E2E_COVERAGE_DIR")
	if capture != "" {
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+capture)
	}
	var logs bytes.Buffer
	cmd.Stdout, cmd.Stderr = &logs, &logs
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var once sync.Once
	stop := func() {
		once.Do(func() {
			if runtime.GOOS == "windows" {
				_ = cmd.Process.Kill()
			} else if err := cmd.Process.Signal(os.Interrupt); err != nil {
				_ = cmd.Process.Kill()
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case err := <-done:
				if err != nil && capture != "" {
					t.Errorf("salida del binario instrumentado: %v", err)
				}
			case <-time.After(15 * time.Second):
				_ = cmd.Process.Kill()
				<-done
				t.Error("el binario no se detuvo a tiempo")
			}
			if t.Failed() {
				t.Logf("stdout API:\n%s", logs.String())
			}
		})
	}
	t.Cleanup(stop)
	s := server{url: "http://" + addr}
	for until := time.Now().Add(12 * time.Second); time.Now().Before(until); {
		if s.request("GET", "/healthz", "", "", false).status == 200 {
			return s, stop
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("el binario no respondió en /healthz")
	return server{}, stop
}

func identity(t *testing.T) (string, func(string, string, string, time.Time) string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	text := func(data []byte) string { return base64.RawURLEncoding.EncodeToString(data) }
	idp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jwks" {
			http.NotFound(w, r)
			return
		}
		if !api.UUIDPattern.MatchString(r.Header.Get(api.RequestIDHeader)) {
			t.Error("la consulta JWKS no propagó X-Request-Id")
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("la carga de JWKS filtró credenciales")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": "e2e", "kty": "RSA", "alg": "RS256", "use": "sig",
			"n": text(key.N.Bytes()), "e": text(big.NewInt(int64(key.E)).Bytes()),
		}}})
	}))
	t.Cleanup(idp.Close)
	sign := func(id, role, aud string, exp time.Time) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": idp.URL, "sub": id, "aud": aud, "exp": exp.Unix(), "perfil": role,
		})
		token.Header["kid"] = "e2e"
		raw, err := token.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	return idp.URL, sign
}

func TestConflictoConcurrente(t *testing.T) {
	db := os.Getenv("TEST_DATABASE_URL")
	bin := os.Getenv("E2E_API_BIN")
	if db == "" {
		t.Skip("requiere TEST_DATABASE_URL sobre una base PostgreSQL de pruebas migrada")
	}
	if bin == "" {
		t.Fatal("falta E2E_API_BIN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, db)
	if err != nil {
		t.Fatal("conexión de prueba inválida")
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal("PostgreSQL de prueba no disponible")
	}
	issuer, sign := identity(t)
	exp := time.Now().Add(time.Minute)
	editor, reader := sign(owner, "editor", audience, exp), sign(owner, "lector", audience, exp)
	other, admin := sign(stranger, "editor", audience, exp), sign(stranger, "administrador", audience, exp)
	s, stop := start(t, bin, db, issuer)
	s.request("GET", "/readyz", "", "", false).check(t, 200)
	s.request("GET", "/api/docs", "", "", false).check(t, 200)
	s.request("GET", "/api/docs/openapi.json", "", "", false).check(t, 200)
	s.request("GET", "/health/live", "", "", false).check(t, 404)
	s.request("PATCH", base, editor, "", false).check(t, 405)
	s.request("GET", base, "", "", false).check(t, 401)
	s.request("GET", base, sign(owner, "editor", "otra-audiencia", exp), "", false).check(t, 401)
	s.request("GET", base, sign(owner, "editor", audience, time.Now().Add(-time.Minute)), "", false).check(t, 401)
	s.request("GET", base, sign(owner, "desconocido", audience, exp), "", false).check(t, 403)
	s.request("GET", base+"?limit=101", editor, "", false).check(t, 400)
	s.request("GET", base+"/invalid", editor, "", false).check(t, 400)
	const body = `{"nombre":"E2E","contenido":{"titulo":"Inicial"}}`
	s.request("POST", base, reader, body, false).check(t, 403)
	s.request("POST", base, editor, `{"nombre":"E2E","contenido":{}}`, false).check(t, 400)
	created := decode[api.Response](t, s.request("POST", base, editor, body, false).check(t, 201))
	if !api.UUIDPattern.MatchString(created.ID) {
		t.Fatal("id de plantilla inválido")
	}
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM auditoria_plantillas WHERE plantilla_id=$1;
			DELETE FROM plantilla_versiones WHERE plantilla_id=$1;
			DELETE FROM plantillas WHERE id=$1`, pgx.QueryExecModeSimpleProtocol, created.ID)
		if err != nil {
			t.Error("limpieza de datos de prueba fallida")
		}
	})
	path := base + "/" + created.ID
	s.request("GET", path, other, "", false).check(t, 404)
	s.request("GET", path, reader, "", true).check(t, 200)
	s.request("GET", path, admin, "", false).check(t, 200)
	results := make(chan response, 2)
	fire := make(chan struct{})
	for range 2 {
		go func() {
			<-fire
			results <- s.request("PUT", path, editor, `{"nombre":"E2E versión 2","contenido":{"titulo":"Actualizado"},"activo":false,"version_esperada":1}`, false)
		}()
	}
	close(fire)
	var statuses []int
	for range 2 {
		r := <-results
		if r.err != nil {
			t.Fatal(r.err)
		}
		statuses = append(statuses, r.status)
	}
	slices.Sort(statuses)
	if !slices.Equal(statuses, []int{200, 409}) {
		t.Fatalf("protección de versión: respuestas %v, esperadas [200 409]", statuses)
	}

	stop()
	s, _ = start(t, bin, db, issuer)
	current := decode[api.Response](t, s.request("GET", path, editor, "", false).check(t, 200))
	var storedCreated, storedUpdated time.Time
	if err := pool.QueryRow(ctx, "SELECT created_at, updated_at FROM plantillas WHERE id=$1", created.ID).
		Scan(&storedCreated, &storedUpdated); err != nil {
		t.Fatal(err)
	}
	versions := decode[api.ListResponse[api.VersionResponse]](t, s.request("GET", path+"/versiones", editor, "", false).check(t, 200))
	before := decode[api.VersionResponse](t, s.request("GET", path+"/versiones/1", editor, "", false).check(t, 200))
	s.request("GET", path+"/versiones/0", editor, "", false).check(t, 400)
	s.request("GET", path+"/versiones/999", editor, "", false).check(t, 404)
	list := decode[api.ListResponse[api.Response]](t, s.request("GET", base, reader, "", false).check(t, 200))
	if !slices.ContainsFunc(list.Data, func(item api.Response) bool { return item.ID == created.ID }) {
		t.Fatal("falta la plantilla en el listado")
	}
	s.request("DELETE", path, admin, "", false).check(t, 204)
	s.request("GET", path, editor, "", false).check(t, 404)
	audits, err := pool.Query(ctx, `SELECT accion, actor_id::text, version, ocurrido_at
		FROM auditoria_plantillas WHERE plantilla_id = $1 ORDER BY id`, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer audits.Close()
	var audit []string
	for audits.Next() {
		var action, actor string
		var version int
		var happened time.Time
		if err := audits.Scan(&action, &actor, &version, &happened); err != nil {
			t.Fatal(err)
		}
		if happened.IsZero() {
			t.Fatal("auditoría sin fecha")
		}
		audit = append(audit, fmt.Sprintf("%s:%s:%d", action, actor, version))
	}
	if err := audits.Err(); err != nil {
		t.Fatal(err)
	}
	var history []int
	for _, v := range versions.Data {
		history = append(history, v.Numero)
	}
	zone := os.Getenv("TZ")
	if zone == "" {
		zone = "UTC"
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	result := struct {
		Statuses []int    `json:"respuestas_concurrentes"`
		Version  int      `json:"version_tras_reinicio"`
		History  []int    `json:"historial"`
		OldTitle any      `json:"contenido_v1"`
		UTC      bool     `json:"fechas_utc_y_mismo_instante"`
		Audit    []string `json:"auditoria"`
	}{
		statuses, current.VersionActual, history, before.Contenido["titulo"],
		current.CreatedAt.Location() == time.UTC && current.UpdatedAt.Location() == time.UTC &&
			current.CreatedAt.Equal(created.CreatedAt) && current.CreatedAt.Equal(storedCreated.In(loc)) &&
			current.UpdatedAt.Equal(storedUpdated.In(loc)),
		audit,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../../e2e/testdata/concurrencia.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(data), bytes.TrimSpace(bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n")))) {
		t.Fatalf("golden distinto\nobtenido:\n%s\nesperado:\n%s", data, want)
	}
}
