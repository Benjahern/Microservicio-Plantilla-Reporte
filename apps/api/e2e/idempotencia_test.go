package e2e

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIdempotenciaPersistente(t *testing.T) {
	db, bin := os.Getenv("TEST_DATABASE_URL"), os.Getenv("E2E_API_BIN")
	if db == "" {
		t.Skip("requiere TEST_DATABASE_URL")
	}
	if bin == "" {
		t.Fatal("falta E2E_API_BIN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, db)
	if err != nil {
		t.Fatal("conexión de prueba inválida")
	}
	t.Cleanup(pool.Close)
	actor := api.NewUUID()
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, err := pool.Exec(cleanup, `DELETE FROM idempotencia_plantillas WHERE actor_id=$1;
			DELETE FROM auditoria_plantillas WHERE actor_id=$1;
			DELETE FROM plantilla_versiones WHERE autor_id=$1;
			DELETE FROM plantillas WHERE propietario_id=$1`, pgx.QueryExecModeSimpleProtocol, actor)
		if err != nil {
			t.Error("limpieza fallida")
		}
	})
	issuer, sign := identity(t)
	token := sign(actor, "editor", audience, time.Now().Add(time.Minute))
	s, stop := start(t, bin, db, issuer)
	session := s.request("POST", "/api/v1/sesiones", token, "", false).check(t, 204)
	cookies := (&http.Response{Header: session.header}).Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatal("atributos de la cookie incorrectos")
	}
	s.request("GET", base, cookies[0].Value, "", true).check(t, 200)

	key := func(values ...string) http.Header { return http.Header{"Idempotency-Key": values} }
	const creation = `{"nombre":"Idempotencia","contenido":{"x":9007199254740993}}`
	s.request("POST", base, token, creation, false, key("duplicada", "duplicada")).check(t, 400)
	s.request("POST", base, token, creation, false, key("clave inválida")).check(t, 400)
	insertKey := key("creacion.1")
	results := make(chan response, 2)
	for range 2 {
		go func() { results <- s.request("POST", base, token, creation, false, insertKey) }()
	}
	a, b := (<-results).check(t, 201), (<-results).check(t, 201)
	if !bytes.Equal(a.body, b.body) {
		t.Fatal("creación repetida no reprodujo la respuesta")
	}
	created := decode[api.Response](t, a)
	failed := s.request("POST", base, token, `{"nombre":"Distinta","contenido":{"x":1}}`, false, insertKey).check(t, 409)
	if decode[api.Problema](t, failed).Code != "idempotency_conflict" {
		t.Fatal("código de conflicto incorrecto")
	}

	path := base + "/" + created.ID
	updateKey := key("edicion.1")
	const stale = `{"nombre":"Idempotencia","contenido":{"x":false},"activo":true,"version_esperada":9}`
	const update = `{"nombre":"Idempotencia","contenido":{"x":true},"activo":true,"version_esperada":1}`
	s.request("PUT", path, token, stale, false, updateKey).check(t, 409)
	updated := s.request("PUT", path, token, update, false, updateKey).check(t, 200)
	stop()
	s, _ = start(t, bin, db, issuer)
	if replay := s.request("PUT", path, token, update, false, updateKey).check(t, 200); !bytes.Equal(updated.body, replay.body) {
		t.Fatal("actualización repetida no reprodujo la respuesta después del reinicio")
	}
	if replay := s.request("POST", base, token, creation, false, insertKey).check(t, 201); !bytes.Equal(a.body, replay.body) {
		t.Fatal("la creación original cambió después de editar y reiniciar")
	}
	for _, check := range []struct {
		sql  string
		want int
	}{
		{"SELECT count(*) FROM plantillas WHERE propietario_id=$1", 1},
		{"SELECT count(*) FROM plantilla_versiones WHERE autor_id=$1", 2},
		{"SELECT count(*) FROM auditoria_plantillas WHERE actor_id=$1", 2},
		{"SELECT count(*) FROM idempotencia_plantillas WHERE actor_id=$1 AND respuesta IS NOT NULL", 2},
	} {
		var count int
		if err := pool.QueryRow(ctx, check.sql, actor).Scan(&count); err != nil || count != check.want {
			t.Fatalf("%s: resultado=%d, esperado=%d (%v)", check.sql, count, check.want, err)
		}
	}
}
