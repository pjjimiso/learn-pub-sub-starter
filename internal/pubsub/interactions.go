package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("Failed to marshal json: %v", err)
	}

	ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)

	return nil
}

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Failed to create channel: %v", err)
	}

	var isDurable bool
	var autoDelete bool
	var exclusive bool

	if queueType == SimpleQueueDurable {
		isDurable = true
		autoDelete = false
		exclusive = false
	} else {
		isDurable = false
		autoDelete = true
		exclusive = true
	}

	queue, err := channel.QueueDeclare(queueName, isDurable, autoDelete, exclusive, false, amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return channel, amqp.Queue{}, fmt.Errorf("Queue declaration failed: %v", err)
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return channel, queue, fmt.Errorf("Queue bind failed: %v", err)
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	c, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	messages, err := c.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			var target T
			err := json.Unmarshal(msg.Body, &target)
			if err != nil {
			}
			response := handler(target)
			switch response {
			case Ack:
				msg.Ack(false)
				log.Println("sent Ack response")
			case NackDiscard:
				msg.Nack(false, false)
				log.Printf("sent NackDiscard response")
			case NackRequeue:
				msg.Nack(false, true)
				log.Printf("sent NackRequeue response")
			}
		}
	}()

	return nil
}
