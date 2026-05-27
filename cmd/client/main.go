package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"

	"github.com/pjjimiso/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/pubsub"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("failed to connect to rabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection successful!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	gs := gamelogic.NewGameState(username)

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open channel: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+fmt.Sprintf(".%s", username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("Pause subscription failed: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, channel),
	)
	if err != nil {
		log.Fatalf("Move subscription failed: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs),
	)
	if err != nil {
		log.Fatalf("War subscription failed: %v", err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			err := gs.CommandSpawn(input)
			if err != nil {
				fmt.Printf("error spawning unit: %v", err)
				continue
			}
		case "move":
			move, err := gs.CommandMove(input)
			if err != nil {
				fmt.Printf("error moving unit: %v", err)
				continue
			}

			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+gs.GetUsername(),
				move,
			)
			fmt.Println("move published successfully...")
		case "status":
			gs.CommandStatus()
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
