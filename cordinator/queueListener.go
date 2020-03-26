package cordinator

import (
	"Distributed-WEB-App/dto"
	"Distributed-WEB-App/qutils"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/streadway/amqp"
)

var url = "amqp://guest:guest@localhost:5672/"

type QueueListener struct {
	con     *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery
	ea      *EventAggregator
}

func NewQueueListener() *QueueListener {
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
		ea:      NewEventAggregator(),
	}

	ql.con, ql.ch = qutils.GetChannel(url)

	return &ql
}

func (ql *QueueListener) ListenForNewSource() {
	q := qutils.GetQueue("", ql.ch, true)

	ql.ch.QueueBind(q.Name, "", "amq.fanout", false, nil)

	msgs, _ := ql.ch.Consume(q.Name, "", true, false, false, false, nil)

	ql.DiscoverSensor()

	for msg := range msgs {
		sourceChan, _ := ql.ch.Consume(string(msg.Body), "", true, false, false, false, nil)

		if ql.sources[string(msg.Body)] == nil {
			ql.sources[string(msg.Body)] = sourceChan

			go ql.AddListener(sourceChan)
		}

	}
}
func (ql *QueueListener) DiscoverSensor() {

	ql.ch.ExchangeDeclare(qutils.SensorDiscoveryExchange, "fanout", false, false, false, false, nil)
	ql.ch.Publish(qutils.SensorDiscoveryExchange, "", false, false, amqp.Publishing{})
}
func (ql *QueueListener) AddListener(msgs <-chan amqp.Delivery) {
	for msg := range msgs {

		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sd := new(dto.SensorMessage)
		d.Decode(sd)

		fmt.Printf("Received message : %v\n", sd)

		ed := EventData{
			Name:      sd.Name,
			Value:     sd.Value,
			TimeStamp: sd.TimeStamp,
		}

		ql.ea.PublishEvent("message received "+msg.RoutingKey, ed)

	}
}
