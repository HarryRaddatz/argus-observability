package docker

import (
	"testing"
	"time"
)

func TestStopDisposition(t *testing.T) {
	cases := []struct {
		state string
		code  int
		oom   bool
		want  string
	}{
		{"running", 0, false, ""},
		{"exited", 0, false, DispositionIntentional},
		{"exited", 143, false, DispositionIntentional},
		{"dead", 137, false, DispositionIntentional},
		{"exited", 1, false, DispositionUnexpected},
		{"exited", 137, true, DispositionOOM},
		{"paused", 0, false, ""},
	}
	for _, c := range cases {
		if got := StopDisposition(c.state, c.code, c.oom); got != c.want {
			t.Errorf("state %s code %d oom %v: got %q, want %q", c.state, c.code, c.oom, got, c.want)
		}
	}
}

func TestMapDockerEventStop(t *testing.T) {
	raw := dockerEvent{Action: "stop", Time: time.Now().Unix()}
	raw.Actor.Attributes = map[string]string{"exitCode": "0"}
	evt, ok := mapDockerEvent("host", "api", raw)
	if !ok || evt.Type != "container.stop" || evt.Payload["cause"] != DispositionIntentional {
		t.Fatalf("stop event: ok=%v type=%s cause=%v", ok, evt.Type, evt.Payload["cause"])
	}

	raw.Action = "die"
	raw.Actor.Attributes["exitCode"] = "1"
	evt, ok = mapDockerEvent("host", "api", raw)
	if !ok || evt.Severity != "warning" || evt.Payload["cause"] != DispositionUnexpected {
		t.Fatalf("die event: ok=%v sev=%s cause=%v", ok, evt.Severity, evt.Payload["cause"])
	}
}
