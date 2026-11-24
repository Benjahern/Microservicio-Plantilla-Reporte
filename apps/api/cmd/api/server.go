package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"github.com/example/microservicio-plantilla-reporte/internal/config"
	"github.com/example/microservicio-plantilla-reporte/internal/httpapi"
	"github.com/example/microservicio-plantilla-reporte/internal/plantilla"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuración inválida", "motivo", err.Error())
		os.Exit(1)
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("DATABASE_URL inválida")
		os.Exit(1)
	}
	poolConfig.ConnConfig.ConnectTimeout = cfg.DatabaseTimeout
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = strconv.FormatInt(max(1, cfg.DatabaseTimeout.Milliseconds()), 10)
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("No se pudo configurar PostgreSQL")
		os.Exit(1)
	}
	defer pool.Close()
	ready := func(parent context.Context) error {
		ctx, cancel := context.WithTimeout(parent, cfg.DatabaseTimeout)
		defer cancel()
		return pool.Ping(ctx)
	}
	repo := plantilla.NewPostgresRepository(pool)
	verifier := auth.NewJWT(cfg.JWKSURL, cfg.Issuer, cfg.Audience, cfg.RoleClaim)
	router := httpapi.NewRouter(verifier, plantilla.NewHandler(plantilla.NewService(repo)), ready, cfg.Environment == "dev")
	if err := serve(cfg.HTTPAddr, httpapi.Observability(router, logger, cfg.Version)); err != nil {
		logger.Error("servidor detenido")
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
