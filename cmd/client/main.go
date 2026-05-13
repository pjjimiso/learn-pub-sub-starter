package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"

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
		fmt.Println("Failed to declare/bind queue:", err)
	}

	gamestate := gamelogic.NewGameState(username)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			err := gamestate.CommandSpawn(input)
			if err != nil {
				fmt.Println("Failed to spawn unit:", err)
				continue
			}
		case "move":
			_, err := gamestate.CommandMove(input)
			if err != nil {
				fmt.Println("Failed to move unit:", err)
				continue
			}
			// TODO: publish move
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}
