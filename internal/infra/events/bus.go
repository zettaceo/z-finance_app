package events

import (
	"sync"
	"time"
)

type Event struct {
	Name       string
	Payload    any
	OccurredAt time.Time
}

type Handler func(Event)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

func (b *Bus) Subscribe(name string, handler Handler) {
	b.mu.Lock()
	b.handlers[name] = append(b.handlers[name], handler)
	b.mu.Unlock()
}

func (b *Bus) Publish(name string, payload any) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[name]...)
	b.mu.RUnlock()
	event := Event{
		Name:       name,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
	}
	for _, handler := range handlers {
		handler(event)
	}
}
