package qutils

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const SensorListQueue = "SensorList"
const SensorDiscoveryExchange = "SensorDiscovery"

func GetChannel(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to rabbitmq server")

	ch, err := conn.Channel()
	failOnError(err, "Failed to connect to channel")

	return conn, ch
}

func GetQueue(name string, ch *amqp.Channel, autoDelete bool) *amqp.Queue {

	q, err := ch.QueueDeclare(name, false, autoDelete, false, false, nil)
	failOnError(err, "Failed  to declare a queue")

	return &q

}

func failOnError(err interface{}, s string) {
	if err != nil {
		log.Fatalf("%s : %s", s, err)
		panic(fmt.Sprintf("%s : %s", s, err))
	}
}
