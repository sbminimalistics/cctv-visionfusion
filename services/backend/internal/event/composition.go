package event

import (
	"sync"
	"time"

	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
)

type CompositionTriggerType string

const (
	CompositionTrigger CompositionTriggerType = "trigger_event"
)

type CompositionTriggerEvent struct {
	uuid          string
	EventType     CompositionTriggerType
	CompositionID string
	Sources       []config.CompositionSourceConfig
}

func (r *CompositionTriggerEvent) GetUUID() string {
	return r.uuid
}

func (r *CompositionTriggerEvent) SetUUID(uuid string) {
	r.uuid = uuid
}

type EventContainer struct {
	event    edge_device.EventPacket
	arrival  time.Time
	expiry   time.Time
	consumed bool
}

type EventComposition struct {
	layout   config.EventCompositionConfig
	events   map[string][]EventContainer
	triggers chan<- CompositionTriggerEvent
	mu       sync.Mutex
}

func (r *EventComposition) AttachChannel(pipe chan<- CompositionTriggerEvent) {
	r.triggers = pipe
}

func (r *EventComposition) TrackEvent(event edge_device.EventPacket) {
	id := event.DeviceID
	if r.events[id] == nil {
		return
	}
	now := time.Now().UTC()
	container := EventContainer{
		event:   event,
		arrival: now,
		expiry:  now.Add(time.Millisecond * time.Duration(r.layout.DefaultEventRetentionMS)),
	}
	r.events[id] = append(r.events[id], container)
	r.checkTrigger()
}

func (r *EventComposition) checkTrigger() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.removeExpiredAndConsumed()

	triggeringContainers := []EventContainer{}
	triggered := false
	for _, src := range r.layout.Sources {
		id := src.DeviceID
		srcEvents := r.events[id]

		found := false
		for _, ev := range srcEvents {
			t := ev.event.Parsed.GetEventType()
			if t == src.EventType {
				found = true
				triggeringContainers = append(triggeringContainers, ev)
				break
			}
		}
		if !found {
			triggered = false
			return
		} else {
			triggered = true
		}
	}

	if triggered == true {
		for _, c := range triggeringContainers {
			c.consumed = true
		}
		r.triggers <- CompositionTriggerEvent{
			EventType:     CompositionTrigger,
			CompositionID: r.layout.Id,
			Sources:       r.layout.Sources,
		}
	}
}

func (r *EventComposition) removeExpiredAndConsumed() {
	now := time.Now().UTC()
	for k, evs := range r.events {
		filtered := evs[:0]
		for _, ev := range evs {
			if !ev.expiry.Before(now) || ev.consumed {
				filtered = append(filtered, ev)
			}
		}

		r.events[k] = filtered
	}
}

func (r *EventComposition) GetId() string {
	return r.layout.Id
}

func NewComposition(cfg config.EventCompositionConfig) *EventComposition {
	ec := EventComposition{
		layout: cfg,
		events: map[string][]EventContainer{},
	}
	for _, src := range cfg.Sources {
		ec.events[src.DeviceID] = []EventContainer{}
	}
	return &ec
}
