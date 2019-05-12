package amqpwrapper

import (
	"log"

	"github.com/streadway/amqp"
)

func ExampleNewManager() {
	_, err := NewManager("amqp://guest:guest@localhost:5672/", amqp.Config{})
	if err != nil {
		log.Fatalf("Error in creating new connection manager: %s", err)
	}
}

func ExampleConnectionManager_CreateChannel() {
	manager, err := NewManager("amqp://guest:guest@localhost:5672/", amqp.Config{})
	if err != nil {
		log.Fatalf("Error in creating new connection manager: %s", err)
	}
	prodChan, err := manager.CreateChannel(Producer)
	if err != nil {
		log.Fatalf("Error in creating new channel: %s", err)
	}
	defer prodChan.Close()
}

func ExampleConnectionManager_GetChannel() {
	manager, err := NewManager("amqp://guest:guest@localhost:5672/", amqp.Config{})
	if err != nil {
		log.Fatalf("Error in creating new connection manager: %s", err)
	}
	prodChan, err := manager.GetChannel("Already Initialized Channel", Producer)
	if err != nil {
		log.Fatalf("Error in getting the channel: %s", err)
	}
	defer prodChan.Close()
}

func ExampleConnectionManager_InitChannel() {
	manager, err := NewManager("amqp://guest:guest@localhost:5672/", amqp.Config{})
	if err != nil {
		log.Fatalf("Error in creating new connection manager: %s", err)
	}
	prodChan, err := manager.CreateChannel(Producer)
	if err != nil {
		log.Fatalf("Error in creating new channel: %s", err)
	}
	initChannel := func(amqpChan *amqp.Channel) (err error) {
		err = amqpChan.ExchangeDeclare("hello", "topic", true, false, false, false, nil)
		return
	}
	args := InitArgs{
		Channel:  prodChan,
		Key:      "New Producer",
		TypeChan: Consumer,
	}
	err = manager.InitChannel(initChannel, args)
	if err != nil {
		log.Fatalf("Error in initializing the channel: %s", err)
	}
}
