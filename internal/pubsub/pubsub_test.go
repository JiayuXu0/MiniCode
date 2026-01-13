package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBroker_SubscribeAndPublish(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := broker.Subscribe(ctx)
	broker.Publish(CreatedEvent, "hello")

	select {
	case event := <-events:
		if event.Type != CreatedEvent {
			t.Errorf("expected CreatedEvent, got %v", event.Type)
		}
		if event.Payload != "hello" {
			t.Errorf("expected 'hello', got %v", event.Payload)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for event")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub1 := broker.Subscribe(ctx)
	sub2 := broker.Subscribe(ctx)
	sub3 := broker.Subscribe(ctx)

	if broker.SubscriberCount() != 3 {
		t.Errorf("expected 3 subscribers, got %d", broker.SubscriberCount())
	}

	broker.Publish(CreatedEvent, 42)

	for i, sub := range []<-chan Event[int]{sub1, sub2, sub3} {
		select {
		case event := <-sub:
			if event.Payload != 42 {
				t.Errorf("subscriber %d: expected 42, got %d", i, event.Payload)
			}
		case <-time.After(time.Second):
			t.Errorf("subscriber %d: timeout", i)
		}
	}
}

func TestBroker_ContextCancel(t *testing.T) {
	broker := NewBroker[string]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	events := broker.Subscribe(ctx)

	if broker.SubscriberCount() != 1 {
		t.Errorf("expected 1 subscriber, got %d", broker.SubscriberCount())
	}

	cancel()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if broker.SubscriberCount() == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if broker.SubscriberCount() != 0 {
		t.Errorf("expected 0 subscribers after cancel, got %d", broker.SubscriberCount())
	}

	select {
	case _, ok := <-events:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for channel close")
	}
}

func TestBroker_NonBlocking(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = broker.Subscribe(ctx)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			broker.Publish(CreatedEvent, i)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Publish blocked")
	}
}

func TestBroker_Concurrent(t *testing.T) {
	broker := NewBroker[int]()
	defer broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const subscribers = 10
	ready := make(chan struct{}, subscribers)
	var subWg sync.WaitGroup

	for i := 0; i < subscribers; i++ {
		subWg.Add(1)
		go func() {
			defer subWg.Done()
			events := broker.Subscribe(ctx)
			ready <- struct{}{}
			for range events {
			}
		}()
	}

	for i := 0; i < subscribers; i++ {
		<-ready
	}

	var pubWg sync.WaitGroup
	for i := 0; i < 10; i++ {
		pubWg.Add(1)
		go func(id int) {
			defer pubWg.Done()
			for j := 0; j < 100; j++ {
				broker.Publish(CreatedEvent, id*100+j)
			}
		}(i)
	}

	pubWg.Wait()
	cancel()

	done := make(chan struct{})
	go func() {
		subWg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("subscribers did not exit")
	}
}
