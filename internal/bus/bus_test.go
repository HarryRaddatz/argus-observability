package bus

import (
	"sync/atomic"
	"testing"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestPublishSubscribe(t *testing.T) {
	b := New()
	var n atomic.Int32
	b.Subscribe(func(model.Event) { n.Add(1) })
	b.Publish(model.Event{Type: "test"})
	if n.Load() != 1 {
		t.Fatalf("got %d", n.Load())
	}
}
