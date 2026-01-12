package event

import (
	"fmt"
	"sync"
)

// Event represents an application event
type Event interface {
	Type() string
}

// EventPublisher publishes events
type EventPublisher struct {
	listeners map[string][]EventListener
	mu        sync.RWMutex
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		listeners: make(map[string][]EventListener),
	}
}

// Publish publishes an event
func (ep *EventPublisher) Publish(event Event) error {
	ep.mu.RLock()
	listeners := ep.listeners[event.Type()]
	ep.mu.RUnlock()

	for _, listener := range listeners {
		if err := listener.OnEvent(event); err != nil {
			return fmt.Errorf("listener error: %v", err)
		}
	}

	return nil
}

// Subscribe subscribes to an event type
func (ep *EventPublisher) Subscribe(eventType string, listener EventListener) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.listeners[eventType] = append(ep.listeners[eventType], listener)
}

// EventListener handles events
type EventListener interface {
	OnEvent(event Event) error
}

// EventListenerFunc is a function-based event listener
type EventListenerFunc func(event Event) error

func (f EventListenerFunc) OnEvent(event Event) error {
	return f(event)
}
