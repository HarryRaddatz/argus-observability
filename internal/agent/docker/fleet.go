package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

type containerInspect struct {
	RestartCount int `json:"RestartCount"`
	State        struct {
		Status    string `json:"Status"`
		Running   bool   `json:"Running"`
		ExitCode  int    `json:"ExitCode"`
		OOMKilled bool   `json:"OOMKilled"`
		Dead      bool   `json:"Dead"`
		Error     string `json:"Error"`
		Health    *struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
}

func (c *Collector) CollectFleet(ctx context.Context) ([]model.ContainerFleetStatus, error) {
	items, err := c.listAllFilteredContainers(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	const workers = 8
	sem := make(chan struct{}, workers)
	var mu sync.Mutex
	var rows []model.ContainerFleetStatus
	var wg sync.WaitGroup

	for _, it := range items {
		wg.Add(1)
		go func(it containerInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			row := c.fleetRowFromContainer(ctx, it, now)
			mu.Lock()
			rows = append(rows, row)
			mu.Unlock()
		}(it)
	}
	wg.Wait()
	return rows, nil
}

func (c *Collector) fleetRowFromContainer(ctx context.Context, it containerInfo, now time.Time) model.ContainerFleetStatus {
	entityUID, _ := c.entityFor(it.Name, it.Labels)
	service := it.Labels["com.docker.compose.service"]
	state := it.State
	if state == "" {
		state = "unknown"
	}
	row := model.ContainerFleetStatus{
		Container:  it.Name,
		EntityUID:  entityUID,
		Service:    service,
		State:      state,
		StatusText: it.Status,
		UpdatedAt:  now,
	}
	ins, err := c.containerInspect(ctx, it.ID)
	if err != nil {
		return row
	}
	row.RestartCount = ins.RestartCount
	row.ExitCode = ins.State.ExitCode
	row.OOMKilled = ins.State.OOMKilled
	if ins.State.Status != "" {
		row.State = ins.State.Status
	}
	if ins.State.Health != nil {
		row.Health = ins.State.Health.Status
	}
	if ins.State.Error != "" {
		row.StatusText = ins.State.Error
	}
	return row
}

func (c *Collector) containerInspect(ctx context.Context, id string) (*containerInspect, error) {
	url := fmt.Sprintf("http://docker/v1.44/containers/%s/json", id)
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
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("docker inspect: %s %s", resp.Status, b)
	}
	var ins containerInspect
	return &ins, json.NewDecoder(resp.Body).Decode(&ins)
}
