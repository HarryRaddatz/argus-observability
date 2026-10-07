package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/agent"
	"github.com/HarryRaddatz/argus-observability/internal/agent/docker"
	"github.com/HarryRaddatz/argus-observability/internal/agent/kernel"
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

	if _, err := cli.Register(ctx, model.Labels{"host": hostID}); err != nil {
		logger.Error("register", "err", err)
		os.Exit(1)
	}
	logger.Info("agent registered", "agent_id", agentID, "hub", hubURL)

	logState := docker.NewLogState()

	go runLogCollector(ctx, logger, collector, cli, logState, logInterval)
	go runFleetCollector(ctx, logger, collector, cli, fleetInterval)
	go runEventStream(ctx, logger, collector, cli)

	if boolEnv("ARGUS_EBPF") {
		startKernelCollector(ctx, logger, collector, cli, hostID,
			durationEnv("ARGUS_EBPF_INTERVAL", 30*time.Second))
	}

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

// startKernelCollector enables the eBPF source. Failing to load is not fatal:
// the agent keeps reporting the log-derived signals instead.
func startKernelCollector(
	ctx context.Context,
	logger *slog.Logger,
	collector *docker.Collector,
	cli *agent.Client,
	hostID string,
	interval time.Duration,
) {
	kcol, err := kernel.Load()
	if err != nil {
		logger.Warn("kernel collector disabled, falling back to log-derived signals", "err", err)
		return
	}
	for _, reason := range kcol.Degraded() {
		logger.Warn("kernel collector partially attached", "detail", reason)
	}
	logger.Info("kernel collector attached", "interval", interval.String())

	go func() {
		defer kcol.Close()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runKernelCollect(ctx, logger, kcol, collector, cli, hostID)
			}
		}
	}()
}

func runKernelCollect(
	ctx context.Context,
	logger *slog.Logger,
	kcol *kernel.Collector,
	collector *docker.Collector,
	cli *agent.Client,
	hostID string,
) {
	flows, err := kcol.Flows()
	if err != nil {
		logger.Warn("read kernel flows", "err", err)
		return
	}
	if len(flows) == 0 {
		return
	}

	ictx, cancel := context.WithTimeout(ctx, 15*time.Second)
	index, err := collector.WorkloadIndex(ictx)
	cancel()
	if err != nil {
		logger.Warn("workload index", "err", err)
		return
	}

	sample := kernel.BuildSample(flows, index, hostID, time.Now().UTC())
	sctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := cli.SendTopology(sctx, sample.Links); err != nil {
		logger.Warn("send topology", "err", err)
	}
	if err := cli.SendMetrics(sctx, sample.Metrics); err != nil {
		logger.Warn("send kernel metrics", "err", err)
	}
	stats := kcol.Stats()
	logger.Info("kernel sample sent",
		"links", len(sample.Links), "metrics", len(sample.Metrics),
		"kernel_events", stats.Events, "map_errors", stats.MapErrors)
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

func boolEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
