package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
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

	queue, err := channel.QueueDeclare(queueName, isDurable, autoDelete, exclusive, false, nil)
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
	handler func(T),
) error {
	c, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveries, err := c.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			var target T
			err := json.Unmarshal(delivery.Body, &target)
			if err != nil {
				fmt.Printf("failed to unmarshal data: %v", err)
				return
			}
			handler(target)
			err = delivery.Ack(false)
			if err != nil {
				fmt.Printf("ack error: %v", err)
			}
		}
	}()

	return nil
}
