package main

import (
	"Distributed-WEB-App/dto"
	"Distributed-WEB-App/qutils"
	"bytes"
	"encoding/gob"
	"flag"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycle/sec")
var max = flag.Float64("max", 5, "maximum value for generated readings")
var min = flag.Float64("min", 1, "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var value = r.Float64()*(*max-*min) + *min

var nom = (*max-*min)/2 + *min

var url = "amqp://guest:guest@localhost:5672/"

func main() {
	flag.Parse()

	con, ch := qutils.GetChannel(url)
	defer con.Close()
	defer ch.Close()

	dataQueue := qutils.GetQueue(*name, ch, false)
	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")

	PublishQueueName(ch)

	discoveryQueue := qutils.GetQueue("", ch, true)
	ch.QueueBind(discoveryQueue.Name, "", qutils.SensorDiscoveryExchange, false, nil)

	go listenForDiscoveryRequest(discoveryQueue.Name, ch)
	signal := time.Tick(dur)

	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(buf)

	for range signal {

		calcValue()
		reading := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			TimeStamp: time.Now(),
		}
		buf.Reset()
		enc = gob.NewEncoder(buf)
		enc.Encode(reading)

		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}

		ch.Publish("", dataQueue.Name, false, false, msg)

		log.Printf("Reading values : %v", value)
	}

}

func listenForDiscoveryRequest(name string, ch *amqp.Channel) {
	msgs, _ := ch.Consume(name, "", true, false, false, false, nil)

	for range msgs {
		PublishQueueName(ch)
	}
}
func PublishQueueName(ch *amqp.Channel) {
	msg := amqp.Publishing{
		Body: []byte(*name),
	}

	ch.Publish("amq.fanout", "", false, false, msg)
}
func calcValue() {

	var maxStep, minStep float64

	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}

	value += r.Float64()*(maxStep-minStep) + minStep

}
