package pubsub

import "context"

// Subscriber is the event subscription interface.
type Subscriber[T any] interface {
	Subscribe(ctx context.Context) <-chan Event[T]
}

// Publisher is the event publishing interface.
type Publisher[T any] interface {
	Publish(eventType EventType, payload T)
}
