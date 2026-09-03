package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/agent"
	"github.com/HarryRaddatz/argus-observability/internal/agent/docker"
	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	hubURL := env("ARGUS_HUB_URL", "http://127.0.0.1:8080")
	token := os.Getenv("ARGUS_AGENT_TOKEN")
	agentID := env("ARGUS_AGENT_ID", hostname())
	hostID := env("ARGUS_HOST_ID", hostname())
	interval := durationEnv("ARGUS_COLLECT_INTERVAL", 15*time.Second)

	collector, err := docker.NewCollector(hostID)
	if err != nil {
		logger.Error("docker collector", "err", err)
		os.Exit(1)
	}
	defer collector.Close()

	cli := agent.NewClient(agent.Config{
		HubURL:     hubURL,
		AgentToken: token,
		AgentID:    agentID,
		HostID:     hostID,
		Interval:   interval,
	}, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if _, err := cli.Register(ctx, model.Labels{"host": hostID}); err != nil {
		logger.Error("register", "err", err)
		os.Exit(1)
	}
	logger.Info("agent registered", "agent_id", agentID, "hub", hubURL)

	ticker := time.NewTicker(interval)
	heartbeat := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer heartbeat.Stop()

	runCollect := func() {
		cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		points, err := collector.Collect(cctx)
		if err != nil {
			logger.Warn("collect", "err", err)
			return
		}
		if err := cli.SendMetrics(cctx, points); err != nil {
			logger.Warn("send metrics", "err", err)
		} else {
			logger.Info("metrics sent", "count", len(points))
		}
	}

	runCollect()

	for {
		select {
		case <-ctx.Done():
			logger.Info("agent stopped")
			return
		case <-ticker.C:
			runCollect()
		case <-heartbeat.C:
			hctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := cli.Heartbeat(hctx); err != nil {
				logger.Warn("heartbeat", "err", err)
			}
			cancel()
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
