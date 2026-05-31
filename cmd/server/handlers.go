package main

import (
	"fmt"

	"github.com/pjjimiso/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/pubsub"
	"github.com/pjjimiso/learn-pub-sub-starter/internal/routing"
)

func handlerWriteLog(gamelog routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(gamelog)
	if err != nil {
		fmt.Printf("error writing log file: %v", err)
		return pubsub.NackDiscard
	}
	return pubsub.Ack
}
