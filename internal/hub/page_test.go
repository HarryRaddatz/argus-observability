package hub

import (
	"net/http"
	"testing"
)

func TestPageWindow(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/logs/search?limit=10&offset=20", nil)
	limit, offset := pageWindow(req, pageDefault)
	if limit != 10 || offset != 20 {
		t.Fatalf("got %d %d", limit, offset)
	}
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/logs/search?limit=9999&offset=-4", nil)
	limit, offset = pageWindow(req, pageDefault)
	if limit != pageMax || offset != 0 {
		t.Fatalf("clamp got %d %d", limit, offset)
	}
}
