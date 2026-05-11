package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"

	"github.com/pjjimiso/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/pubsub"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Println("Failed to create rabbitMQ connection:", err)
	}
	defer conn.Close()
	fmt.Println("Connection successful!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Invalid username:", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+fmt.Sprintf(".%s", username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		fmt.Println("Failed to declare/bind queue: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Shutting down client...")
	conn.Close()
}
