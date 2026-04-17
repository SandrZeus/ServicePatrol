package events

import (
	"log"
	"sync"
)

type Bus struct {
	mu          sync.Mutex
	subscribers []chan Event
}

func NewBus() *Bus {
	return &Bus{}
}

func (b *Bus) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 32)
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()

	unsubscibe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		for i, sub := range b.subscribers {
			if sub == ch {
				b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
				close(ch)
				return
			}
		}
	}
	return ch, unsubscibe
}

func (b *Bus) Publish(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			log.Printf("events: subsciber bugger full, dropping event type=%s target=%d", event.Type, event.TargetID)
		}
	}
}
