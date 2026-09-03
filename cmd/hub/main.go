package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/hub"
	"github.com/HarryRaddatz/argus-observability/internal/store/sqlite"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	addr := env("ARGUS_HUB_ADDR", ":8080")
	dbPath := env("ARGUS_STORE_PATH", "./data/argus.db")
	token := os.Getenv("ARGUS_AGENT_TOKEN")

	if err := os.MkdirAll("./data", 0o755); err != nil {
		logger.Error("mkdir data", "err", err)
		os.Exit(1)
	}

	st, err := sqlite.Open(dbPath)
	if err != nil {
		logger.Error("open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	eventBus := bus.New()
	srv := hub.New(hub.Config{
		Addr:             addr,
		AgentToken:       token,
		RetentionLogs:    durationEnv("ARGUS_RETENTION_LOGS", 7*24*time.Hour),
		RetentionMetrics: durationEnv("ARGUS_RETENTION_METRICS", 30*24*time.Hour),
		RetentionEvents:  durationEnv("ARGUS_RETENTION_EVENTS", 30*24*time.Hour),
		PurgeInterval:    durationEnv("ARGUS_PURGE_INTERVAL", time.Hour),
		PurgeTimeout:     durationEnv("ARGUS_PURGE_TIMEOUT", 5*time.Second),
	}, st, eventBus, logger)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("hub listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	logger.Info("hub stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
