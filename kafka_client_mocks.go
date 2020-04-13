package main

import (
	"github.com/Financial-Times/kafka-client-go/kafka"
)

type mockConsumer struct {
	message kafka.FTMessage
	err     error
}

func (mc mockConsumer) StartListening(messageHandler func(message kafka.FTMessage) error) {
	_ = messageHandler(mc.message)
}

func (mc mockConsumer) Shutdown() {
}

func (mc mockConsumer) ConnectivityCheck() error {
	return mc.err
}
