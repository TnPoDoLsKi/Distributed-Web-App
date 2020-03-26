package cordinator

import (
	"time"
)

type EventAggregator struct {
	listners map[string]func(EventData)
}

func NewEventAggregator() *EventAggregator {

	ea := EventAggregator{
		listners: make(map[string]func(EventData)),
	}

	return &ea
}

func (ea *EventAggregator) AddListener(name string, f func(data EventData)) {
	ea.listners[name] = f
}

func (ea *EventAggregator) PublishEvent(name string, eventdata EventData) {
	if ea.listners[name] != nil {
		for _, r := range ea.listners {
			r(eventdata)
		}
	}
}

type EventData struct {
	Name      string
	Value     float64
	TimeStamp time.Time
}
