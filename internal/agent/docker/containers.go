package docker

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

type containerInfo struct {
	ID     string
	Name   string
	Labels map[string]string
	State  string
	Status string
}

func (c *Collector) listFilteredContainers(ctx context.Context) ([]containerInfo, error) {
	return c.listFilteredContainersAll(ctx, false)
}

func (c *Collector) listAllFilteredContainers(ctx context.Context) ([]containerInfo, error) {
	return c.listFilteredContainersAll(ctx, true)
}

func (c *Collector) listFilteredContainersAll(ctx context.Context, all bool) ([]containerInfo, error) {
	containers, err := c.listContainers(ctx, all)
	if err != nil {
		return nil, err
	}
	prefix := os.Getenv("ARGUS_NAME_PREFIX")
	var out []containerInfo
	for _, ctr := range containers {
		name := ctr.primaryName()
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		out = append(out, containerInfo{
			ID:     ctr.ID,
			Name:   name,
			Labels: ctr.Labels,
			State:  ctr.State,
			Status: ctr.Status,
		})
	}
	return out, nil
}

func (c *Collector) entityFor(name string, labels map[string]string) (string, model.Labels) {
	entityUID := fmt.Sprintf("docker:%s:%s", c.hostID, name)
	return entityUID, entityLabels(c.hostID, name, labels)
}
