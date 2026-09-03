package docker

import "github.com/HarryRaddatz/argus-observability/internal/model"

func entityLabels(hostID, name string, dockerLabels map[string]string) model.Labels {
	lbl := model.Labels{
		"host": hostID, "runtime": "docker", "container": name,
	}
	if project := dockerLabels["com.docker.compose.project"]; project != "" {
		lbl["stack"] = project
	}
	if svc := dockerLabels["com.docker.compose.service"]; svc != "" {
		lbl["service"] = svc
	}
	return lbl
}

func eventLabels(hostID, name string, attrs map[string]string) model.Labels {
	return entityLabels(hostID, name, attrs)
}
