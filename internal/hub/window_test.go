package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestWindowDuration(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/series?since=15m", nil)
	from, until, err := requestWindow(r, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !until.IsZero() {
		t.Fatalf("open window until = %s", until)
	}
	if time.Since(from) < 14*time.Minute || time.Since(from) > 16*time.Minute {
		t.Fatalf("from = %s", from)
	}
}

func TestRequestWindowClosed(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/series?since=2026-10-06T14:00:00Z..2026-10-06T15:00:00Z", nil)
	from, until, err := requestWindow(r, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !from.Equal(time.Date(2026, 10, 6, 14, 0, 0, 0, time.UTC)) {
		t.Fatalf("from = %s", from)
	}
	if !until.Equal(time.Date(2026, 10, 6, 15, 0, 0, 0, time.UTC)) {
		t.Fatalf("until = %s", until)
	}
}

func TestRequestWindowRejectsWideRange(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/series?since=2026-10-01T00:00:00Z..2026-10-06T00:00:00Z", nil)
	if _, _, err := requestWindow(r, time.Hour); err == nil {
		t.Fatal("expected error")
	}
}
