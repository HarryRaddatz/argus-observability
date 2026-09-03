package bus

import (
	"sync"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

type Handler func(model.Event)

type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
}

func New() *Bus {
	return &Bus{}
}

func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, h)
}

func (b *Bus) Publish(evt model.Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers...)
	b.mu.RUnlock()
	for _, h := range handlers {
		h(evt)
	}
}
