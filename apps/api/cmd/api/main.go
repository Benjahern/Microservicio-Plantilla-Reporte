package main

import (
	"context"
	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"github.com/example/microservicio-plantilla-reporte/internal/config"
	"github.com/example/microservicio-plantilla-reporte/internal/database"
	"github.com/example/microservicio-plantilla-reporte/internal/httpapi"
	"github.com/example/microservicio-plantilla-reporte/internal/plantilla"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	client, err := database.Connect(ctx, cfg)
	if err != nil {
		slog.Error("conexión MongoDB fallida", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())
	repo := plantilla.NewMongoRepository(client, cfg.MongoDatabase)
	idxCtx, cancel := context.WithTimeout(ctx, cfg.MongoTimeout)
	err = repo.EnsureIndexes(idxCtx)
	cancel()
	if err != nil {
		slog.Error("creación de índices fallida", "error", err)
		os.Exit(1)
	}
	svc := plantilla.NewService(repo)
	handler := plantilla.NewHandler(svc)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.NewRouter(auth.NewRegistry(cfg.Identities), handler, func() error {
		c, cancel := context.WithTimeout(context.Background(), cfg.MongoTimeout)
		defer cancel()
		return client.Ping(c, nil)
	}), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		slog.Info("API escuchando", "addr", cfg.HTTPAddr)
		if e := server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			slog.Error("servidor detenido", "error", e)
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		slog.Error("apagado HTTP fallido", "error", err)
	}
}
