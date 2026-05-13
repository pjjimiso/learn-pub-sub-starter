package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/pjjimiso/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/pubsub"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril server...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Println("Failed to create rabbitMQ connection:", err)
	}
	defer conn.Close()
	fmt.Println("Connection successful!")

	// Open channel
	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a new channel:", err)
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			fmt.Println("Sending Pause message...")
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				fmt.Println("Could not publish message:", err)
			}
		case "resume":
			fmt.Println("Sending Resume message...")
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				fmt.Println("Could not publish message:", err)
			}
		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Command not found")
		}
	}
}
