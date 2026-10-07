package agent

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	retryAttempts = 5
	retryBase     = 250 * time.Millisecond
	retryCap      = 8 * time.Second
)

// backoffDelay is the wait before retry n (1 → ~250ms, 2 → ~500ms, …),
// capped, with ±25% jitter so agents do not retry in lockstep.
func backoffDelay(retry int, jitter func(int64) int64) time.Duration {
	if retry < 1 {
		retry = 1
	}
	shift := retry - 1
	if shift > 16 {
		shift = 16
	}
	d := retryBase << shift
	if d > retryCap || d <= 0 {
		d = retryCap
	}
	span := int64(d / 4)
	if span == 0 || jitter == nil {
		return d
	}
	return d + time.Duration(jitter(span*2+1)-span)
}

func waitContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type hubStatusError struct {
	code   int
	status string
	detail string
}

func (e *hubStatusError) Error() string {
	if e.detail == "" {
		return "hub " + e.status
	}
	return "hub " + e.status + ": " + e.detail
}

func retryable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	var status *hubStatusError
	if errors.As(err, &status) {
		return status.code == http.StatusTooManyRequests || status.code >= 500
	}
	return true
}
