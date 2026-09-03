package fleet

import (
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// BuildResponse assembles fleet status for API consumers.
func BuildResponse(containers []model.ContainerFleetStatus, events model.FleetEventStats) model.FleetStatusResponse {
	if containers == nil {
		containers = []model.ContainerFleetStatus{}
	}
	var updated time.Time
	for _, c := range containers {
		if !c.UpdatedAt.IsZero() && c.UpdatedAt.After(updated) {
			updated = c.UpdatedAt
		}
	}
	if updated.IsZero() {
		updated = time.Now().UTC()
	}
	return model.FleetStatusResponse{
		UpdatedAt:  updated,
		Summary:    Summarize(containers),
		Services:   ServiceReplicas(containers),
		Containers: containers,
		Events24h:  events,
	}
}

// Summarize counts container states across the fleet snapshot.
func Summarize(containers []model.ContainerFleetStatus) model.FleetSummary {
	var s model.FleetSummary
	for _, c := range containers {
		s.TotalRestartCount += c.RestartCount
		switch strings.ToLower(c.State) {
		case "running":
			s.Running++
		case "exited":
			s.Exited++
		case "restarting":
			s.Restarting++
		case "dead":
			s.Dead++
		}
		if strings.EqualFold(c.Health, "unhealthy") {
			s.Unhealthy++
		}
	}
	s.ReplicasTotal = len(containers)
	s.ReplicasUp = s.Running + s.Restarting
	return s
}

// ServiceReplicas groups containers by compose service name.
func ServiceReplicas(containers []model.ContainerFleetStatus) []model.ServiceReplicaStatus {
	byService := map[string]*model.ServiceReplicaStatus{}
	for _, c := range containers {
		svc := c.Service
		if svc == "" {
			svc = c.Container
		}
		row, ok := byService[svc]
		if !ok {
			row = &model.ServiceReplicaStatus{Service: svc}
			byService[svc] = row
		}
		row.ReplicasTotal++
		state := strings.ToLower(c.State)
		if state == "running" || state == "restarting" {
			row.ReplicasUp++
		}
		if state == "restarting" {
			row.Restarting++
		}
		if strings.EqualFold(c.Health, "unhealthy") {
			row.Unhealthy++
		}
	}
	out := make([]model.ServiceReplicaStatus, 0, len(byService))
	for _, row := range byService {
		out = append(out, *row)
	}
	return out
}
