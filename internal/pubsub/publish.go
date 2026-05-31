package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	err := encoder.Encode(val)
	if err != nil {
		return fmt.Errorf("gob encoding failed: %v", err)
	}

	ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        b.Bytes(),
		},
	)

	return nil
}
