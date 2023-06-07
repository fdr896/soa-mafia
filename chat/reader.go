package chat

import (
	amqp "github.com/rabbitmq/amqp091-go"
	zlog "github.com/rs/zerolog/log"
)

func (cs *ChatServer) StartReadFromChat() (<-chan amqp.Delivery, error) {
	zlog.Info().Str("queue", cs.queue.Name).Msg("start consuming")

	msgsChan, err := cs.ch.Consume(
		cs.queue.Name, // queue
		"",            // consumer
		true,          // autoAck
		false,         // exclusive
		false,         // noLocal
		false,         // noWait
		nil,           // args
	)
	if err != nil {
		return nil, err
	}

	return msgsChan, nil
}
