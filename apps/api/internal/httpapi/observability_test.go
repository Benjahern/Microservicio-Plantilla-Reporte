package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogDeSolicitud(t *testing.T) {
	var stdout bytes.Buffer
	h := Observability(router(false), slog.New(slog.NewJSONHandler(&stdout, nil)), "prueba")
	req := httptest.NewRequest("POST", base+"?privado=consulta-privada", strings.NewReader("cuerpo-privado"))
	req.Header.Set("Authorization", "Bearer credencial-privada")
	req.Header.Set("Cookie", "sesion=cookie-privada")
	req.Header.Set("X-Request-Id", owner)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 401 || w.Header().Get("X-Request-Id") != owner {
		t.Fatal("estado o request ID incorrecto")
	}
	var fields map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for key, expected := range map[string]any{
		"servicio": "plantillas-reporte", "version": "prueba", "request_id": owner,
		"metodo": "POST", "ruta": base, "estado": float64(401),
	} {
		if fields[key] != expected {
			t.Errorf("%s = %v, esperado %v", key, fields[key], expected)
		}
	}
	if _, ok := fields["duracion_ms"].(float64); !ok {
		t.Error("duracion_ms ausente")
	}
	for _, s := range []string{"consulta-privada", "cuerpo-privado", "credencial-privada", "cookie-privada"} {
		if strings.Contains(stdout.String(), s) {
			t.Error("se registró un dato privado de la petición")
		}
	}
}

func TestSesionCookieRechazaOrigenCruzado(t *testing.T) {
	req := httptest.NewRequest("DELETE", "https://web.example"+base+"/"+owner, nil)
	req.AddCookie(&http.Cookie{Name: "sesion", Value: "editor", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	req.Header.Set("Origin", "https://otro.example")
	w := httptest.NewRecorder()
	router(false).ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("estado = %d, esperado 403", w.Code)
	}
}
