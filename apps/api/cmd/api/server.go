package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"github.com/example/microservicio-plantilla-reporte/internal/config"
	"github.com/example/microservicio-plantilla-reporte/internal/httpapi"
	"github.com/example/microservicio-plantilla-reporte/internal/plantilla"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}
	repo := plantilla.NewMockRepository()
	router := httpapi.NewRouter(auth.NewRegistry(cfg.Identities), plantilla.NewHandler(plantilla.NewService(repo)),
		func(context.Context) error { return nil }, cfg.Environment == "dev")
	if err := serve(cfg.HTTPAddr, router); err != nil {
		slog.Error("servidor detenido", "error", err)
		os.Exit(1)
	}
}

func serve(addr string, router http.Handler) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	result := make(chan error, 1)
	go func() { result <- server.ListenAndServe() }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
