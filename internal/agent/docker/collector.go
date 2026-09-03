package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

const defaultSocket = "/var/run/docker.sock"

type Collector struct {
	client *http.Client
	hostID string
}

func NewCollector(hostID string) (*Collector, error) {
	socket := os.Getenv("DOCKER_HOST")
	if socket == "" {
		socket = "unix://" + defaultSocket
	}
	if !strings.HasPrefix(socket, "unix://") {
		return nil, fmt.Errorf("only unix socket supported, got %q", socket)
	}
	path := strings.TrimPrefix(socket, "unix://")
	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", path)
		},
	}
	if hostID == "" {
		hostID, _ = os.Hostname()
	}
	return &Collector{
		client: &http.Client{Transport: tr, Timeout: 30 * time.Second},
		hostID: hostID,
	}, nil
}

func (c *Collector) Close() error {
	return nil
}

type containerSummary struct {
	ID    string            `json:"Id"`
	Names []string          `json:"Names"`
	Labels map[string]string `json:"Labels"`
}

type statsResponse struct {
	CPUStats    cpuStats    `json:"cpu_stats"`
	PreCPUStats cpuStats    `json:"precpu_stats"`
	MemoryStats memoryStats `json:"memory_stats"`
}

type cpuStats struct {
	CPUUsage struct {
		TotalUsage uint64   `json:"total_usage"`
		PercpuUsage []uint64 `json:"percpu_usage"`
	} `json:"cpu_usage"`
	SystemUsage uint64 `json:"system_cpu_usage"`
	OnlineCPUs  uint32 `json:"online_cpus"`
}

type memoryStats struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

func (c *Collector) Collect(ctx context.Context) ([]model.MetricPoint, error) {
	containers, err := c.listContainers(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var points []model.MetricPoint
	for _, ctr := range containers {
		name := ctr.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		stats, err := c.containerStats(ctx, ctr.ID)
		if err != nil {
			continue
		}
		entityUID := fmt.Sprintf("docker:%s:%s", c.hostID, name)
		labels := model.Labels{
			"host": c.hostID, "runtime": "docker", "container": name,
		}
		if svc := ctr.Labels["com.docker.compose.service"]; svc != "" {
			labels["service"] = svc
		}
		points = append(points,
			model.MetricPoint{MetricName: "cpu.usage", TS: now, Value: cpuPercent(stats), EntityUID: entityUID, Labels: labels},
			model.MetricPoint{MetricName: "memory.usage", TS: now, Value: float64(stats.MemoryStats.Usage), EntityUID: entityUID, Labels: labels},
			model.MetricPoint{MetricName: "memory.limit", TS: now, Value: float64(stats.MemoryStats.Limit), EntityUID: entityUID, Labels: labels},
		)
	}
	return points, nil
}

func (c *Collector) listContainers(ctx context.Context) ([]containerSummary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/v1.44/containers/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("docker list: %s %s", resp.Status, b)
	}
	var out []containerSummary
	return out, json.NewDecoder(resp.Body).Decode(&out)
}

func (c *Collector) containerStats(ctx context.Context, id string) (*statsResponse, error) {
	url := fmt.Sprintf("http://docker/v1.44/containers/%s/stats?stream=false", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("docker stats: %s", resp.Status)
	}
	var stats statsResponse
	return &stats, json.NewDecoder(resp.Body).Decode(&stats)
}

func cpuPercent(v *statsResponse) float64 {
	if v == nil || v.PreCPUStats.SystemUsage == 0 || v.CPUStats.SystemUsage == 0 {
		return 0
	}
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
	if sysDelta <= 0 {
		return 0
	}
	online := v.CPUStats.OnlineCPUs
	if online == 0 {
		online = uint32(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if online == 0 {
		online = 1
	}
	return (cpuDelta / sysDelta) * float64(online) * 100
}
