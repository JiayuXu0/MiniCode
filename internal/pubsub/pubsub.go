package pubsub

import (
	"context"
	"sync"
)

// EventType defines the event type.
type EventType int

const (
	CreatedEvent EventType = iota
	UpdatedEvent
	DeletedEvent
)

// String returns a string representation of the event type.
func (e EventType) String() string {
	switch e {
	case CreatedEvent:
		return "created"
	case UpdatedEvent:
		return "updated"
	case DeletedEvent:
		return "deleted"
	default:
		return "unknown"
	}
}

// Event is a generic event payload wrapper.
type Event[T any] struct {
	Type    EventType
	Payload T
}

const (
	defaultBufferSize = 100
	defaultMaxEvents  = 1000
)

// Broker is a generic pubsub broker.
type Broker[T any] struct {
	mu        sync.RWMutex
	subs      map[chan Event[T]]struct{}
	closed    bool
	bufSize   int
	maxEvents int
}

// NewBroker creates a new broker with default settings.
func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		subs:      make(map[chan Event[T]]struct{}),
		bufSize:   defaultBufferSize,
		maxEvents: defaultMaxEvents,
	}
}

// Subscribe registers a new subscriber and returns its event channel.
// The channel is closed when the context is canceled or the broker shuts down.
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan Event[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		ch := make(chan Event[T])
		close(ch)
		return ch
	}

	ch := make(chan Event[T], b.bufSize)
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.unsubscribe(ch)
	}()

	return ch
}

// Publish publishes an event to all subscribers.
// Non-blocking: if a subscriber is slow, the event is dropped.
func (b *Broker[T]) Publish(eventType EventType, payload T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	event := Event[T]{
		Type:    eventType,
		Payload: payload,
	}

	for ch := range b.subs {
		select {
		case ch <- event:
		default:
		}
	}
}

// Shutdown closes the broker and all subscriber channels.
func (b *Broker[T]) Shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	for ch := range b.subs {
		close(ch)
	}
	b.subs = make(map[chan Event[T]]struct{})
}

// SubscriberCount returns the current subscriber count.
func (b *Broker[T]) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}

func (b *Broker[T]) unsubscribe(ch chan Event[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
}
