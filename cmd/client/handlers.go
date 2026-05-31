package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"

	"github.com/pjjimiso/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/pubsub"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		username := gs.GetUsername()
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			err := PublishGameLog(
				ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+username,
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    username,
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := PublishGameLog(
				ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+username,
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    username,
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := PublishGameLog(
				ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+username,
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
					Username:    username,
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("invalid war outcome")
			return pubsub.NackDiscard
		}
	}
}

func PublishGameLog(ch *amqp.Channel, exchange, key string, log routing.GameLog) error {
	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, key, log)
	if err != nil {
		return fmt.Errorf("error publishing game log: %v", err)
	}
	return nil
}
