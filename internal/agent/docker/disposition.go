package docker

import "strings"

const (
	// DispositionIntentional is a stop the operator asked for: exit 0, or a
	// stop signal, and not an out-of-memory kill.
	DispositionIntentional = "intentional"
	// DispositionUnexpected is a process that exited on its own with a failure code.
	DispositionUnexpected = "unexpected"
	// DispositionOOM is a kill by the memory controller.
	DispositionOOM = "oom"
)

// StopDisposition classifies a container that is not running.
// A running, restarting, or paused container has no stop disposition.
func StopDisposition(state string, exitCode int, oom bool) string {
	switch strings.ToLower(state) {
	case "exited", "dead":
	default:
		return ""
	}
	if oom {
		return DispositionOOM
	}
	// 0: clean exit. 130: SIGINT. 137: SIGKILL after a stop timeout or docker kill.
	// 143: SIGTERM, the signal docker stop sends first.
	switch exitCode {
	case 0, 130, 137, 143:
		return DispositionIntentional
	default:
		return DispositionUnexpected
	}
}
