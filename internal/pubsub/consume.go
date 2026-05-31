package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

type AckType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
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
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("failed to unmarshal json: %v", err)
				continue
			}
			response := handler(target)
			switch response {
			case Ack:
				msg.Ack(false)
				log.Printf("sent Ack response")
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(data []byte) (T, error) {
			var target T
			err := json.Unmarshal(data, &target)
			return target, err
		},
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(data []byte) (T, error) {
			var target T
			decoder := gob.NewDecoder(bytes.NewBuffer(data))
			err := decoder.Decode(&target)
			return target, err
		},
	)
}

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
