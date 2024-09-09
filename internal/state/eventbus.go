package state

import (
	"sync"
)

type EventBus struct {
	subscribers map[string][]chan interface{}
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan interface{}),
	}
}

// TODO enum for eventType
func (eb *EventBus) Subscribe(eventType string, ch chan interface{}) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
}

func (eb *EventBus) Publish(eventType string, data interface{}) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subscribers, ok := eb.subscribers[eventType]
	if !ok {
		return
	}

	for i := 0; i < len(subscribers); i++ {
		ch := subscribers[i]
		select {
		case ch <- data:
			// Success
		default:
			// If cannot receive or closed, remove the subscriber
			eb.mu.Lock()
			if i < len(eb.subscribers[eventType])-1 {
				eb.subscribers[eventType] = append(eb.subscribers[eventType][:i], eb.subscribers[eventType][i+1:]...)
			} else {
				eb.subscribers[eventType] = eb.subscribers[eventType][:i]
			}
			eb.mu.Unlock()

			if i > 0 {
				i--
			}
		}
	}
}

func (eb *EventBus) Unsubscribe(eventType string, ch chan interface{}) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscribers, ok := eb.subscribers[eventType]
	if !ok {
		return
	}

	for i, subscriber := range subscribers {
		if subscriber == ch {
			if i == len(subscribers)-1 {
				eb.subscribers[eventType] = subscribers[:i]
			} else {
				eb.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
			}
			break
		}
	}
	if len(eb.subscribers[eventType]) == 0 {
		delete(eb.subscribers, eventType)
	}
}
