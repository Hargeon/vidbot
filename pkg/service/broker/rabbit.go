// Package broker uses for implementation message broker
package broker

import (
	"github.com/streadway/amqp"
)

// Rabbit represent rabbitmq client
type Rabbit struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue

	url string
}

// NewRabbit initialize Rabbit
func NewRabbit(url string) *Rabbit {
	return &Rabbit{
		url: url,
	}
}

// Connect to rabbit, create chan, declare queue
func (r *Rabbit) Connect(queueName string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(r.url)

	if err != nil {
		return nil, err
	}

	r.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	r.ch = ch

	q, err := ch.QueueDeclare(
		queueName,
		true, // message will not lose if rabbit crashed
		false,
		false,
		false,
		nil)
	if err != nil {
		return nil, err
	}

	r.q = q

	return conn, err
}

// Consume return channel which read messages from rabbit
func (r *Rabbit) Consume() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.q.Name,
		"",
		false, // needs to mark a message was processed
		false,
		false,
		false,
		nil)

	return msgs, err
}
