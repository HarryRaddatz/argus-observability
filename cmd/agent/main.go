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
	logInterval := durationEnv("ARGUS_LOG_INTERVAL", 30*time.Second)
	fleetInterval := durationEnv("ARGUS_FLEET_INTERVAL", 60*time.Second)

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

	if _, err := cli.Register(ctx, model.Labels{
		"host":  hostID,
		"stack": "venuz",
	}); err != nil {
		logger.Error("register", "err", err)
		os.Exit(1)
	}
	logger.Info("agent registered", "agent_id", agentID, "hub", hubURL)

	logState := docker.NewLogState()

	go runLogCollector(ctx, logger, collector, cli, logState, logInterval)
	go runFleetCollector(ctx, logger, collector, cli, fleetInterval)
	go runEventStream(ctx, logger, collector, cli)

	ticker := time.NewTicker(interval)
	heartbeat := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer heartbeat.Stop()

	runCollect := func() {
		cctx, cancel := context.WithTimeout(ctx, 120*time.Second)
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

func runLogCollector(
	ctx context.Context,
	logger *slog.Logger,
	collector *docker.Collector,
	cli *agent.Client,
	logState *docker.LogState,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	run := func() {
		lctx, cancel := context.WithTimeout(ctx, 90*time.Second)
		entries, err := collector.CollectLogs(lctx, logState)
		cancel()
		if err != nil {
			logger.Warn("collect logs", "err", err)
			return
		}
		if len(entries) == 0 {
			return
		}
		sctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := cli.SendLogs(sctx, entries); err != nil {
			logger.Warn("send logs", "err", err)
			return
		}
		logger.Info("logs sent", "count", len(entries))
	}

	run()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func runFleetCollector(
	ctx context.Context,
	logger *slog.Logger,
	collector *docker.Collector,
	cli *agent.Client,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	run := func() {
		fctx, cancel := context.WithTimeout(ctx, 120*time.Second)
		rows, err := collector.CollectFleet(fctx)
		cancel()
		if err != nil {
			logger.Warn("collect fleet", "err", err)
			return
		}
		if len(rows) == 0 {
			return
		}
		sctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := cli.SendFleet(sctx, rows); err != nil {
			logger.Warn("send fleet", "err", err)
			return
		}
		logger.Info("fleet sent", "count", len(rows))
	}

	run()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func runEventStream(ctx context.Context, logger *slog.Logger, collector *docker.Collector, cli *agent.Client) {
	err := collector.StreamEvents(ctx, func(evt model.Event) error {
		sctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := cli.SendEvent(sctx, evt); err != nil {
			return err
		}
		logger.Info("event sent", "type", evt.Type, "entity", evt.EntityUID)
		return nil
	})
	if err != nil && ctx.Err() == nil {
		logger.Warn("event stream stopped", "err", err)
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
