package event

import (
	"context"
	"fmt"
	"sync"

	"sbminimalistics.com/vms/internal/cam"
)

// CommsEvent represents the core event payload passed around the broker
type CommsEvent struct {
	ID        string
	CameraID  string
	EventType string
	Payload   interface{}
}

type Broker struct {
	ch chan CommsEvent
}

func NewBroker(bufferSize int) *Broker {
	return &Broker{
		ch: make(chan CommsEvent, bufferSize),
	}
}

func (b *Broker) Publish(ev CommsEvent) {
	b.ch <- ev
}

func (b *Broker) StartConsumer(ctx context.Context, wg *sync.WaitGroup, alertHandler func(CommsEvent)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-b.ch:
				if !ok {
					return
				}
				alertHandler(ev)
			}
		}
	}()
}

func (b *Broker) OrchestrateStream(ctx context.Context, wg *sync.WaitGroup, stream *cam.EventNotificationAlert) {
	wg.Add(1)

	go func(s *cam.EventNotificationAlert) {
		defer wg.Done()
		fmt.Printf("Orchestrator hooked into alert feed for: %s\n", s.Name)

		// consume messages from the new getter method channel
		for data := range s.Events() {

			// output message directly to stdio from the broker context
			fmt.Printf("@broker event from %s cam\n", s.Name)

			// wrap and route it onward to central system for the consumer to catch
			select {
			case <-ctx.Done():
				return
			case b.ch <- CommsEvent{
				CameraID:  s.Name,
				EventType: data.EventType,
				Payload:   data.EventDescription,
			}:
			}
		}
	}(stream)
}
