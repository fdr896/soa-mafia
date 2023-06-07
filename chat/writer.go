package chat

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	zlog "github.com/rs/zerolog/log"
)

type ChatMessage struct {
	Username string    `json:"username"`
	Message  string    `json:"message"`
	SendTime time.Time `json:"send_time"`
}

func (cs *ChatServer) WriteToChat(msg *ChatMessage, chat string) error {
	if err := cs.isValidateChat(chat); err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(errBadMessage, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	zlog.Info().Str("chat", chat).Msg("publishing msg")
	return cs.ch.PublishWithContext(
		ctx,   // ctx (timeout after some time)
		chat,  // exchange
		"",    // routing key (falout ignores routing key)
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body: msgBytes,
		},     // msg
	)
}
