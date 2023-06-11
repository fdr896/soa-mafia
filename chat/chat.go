package chat

import (
	"common"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	zlog "github.com/rs/zerolog/log"
)

const (
	DAY_CHAT = "day_chat"
	NIGHT_CHAT = "night_chat"
)

type ChatServer struct {
	username string
	sessionId string
	connParams *common.RabbitmqConnectionParams

	conn *amqp.Connection
	ch *amqp.Channel
	queue amqp.Queue
}

func NewChatServer(username, sessionId string, connParams *common.RabbitmqConnectionParams) *ChatServer {
	return &ChatServer{
		username: username,
		sessionId: sessionId,
		connParams: connParams,
	}
}

func (cs *ChatServer) GetSessionChatName(baseChat string) string {
	return fmt.Sprintf("%s_%s", baseChat, cs.sessionId)
}

func (cs *ChatServer) GetChatQueueName() string {
	return fmt.Sprintf("%s_%s", cs.sessionId, cs.username)
}

func (cs *ChatServer) StartChat() error {
	connUrl := common.GetRabbitmqConnectionUrl(cs.connParams)
	zlog.Info().Str("url", connUrl).Msg("connecting to rabbitmq")

	conn, err := amqp.Dial(connUrl)
	if err != nil {
		return err
	}
	cs.conn = conn
	zlog.Info().Msg("connection establised")

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	cs.ch = ch
	zlog.Info().Msg("chan created")

	if err := cs.declareDayChatExchange(); err != nil {
		zlog.Error().Err(err).Msg("failed to declare day chat exchange")
		return  err
	}
	if err := cs.declareNightChatExchange(); err != nil {
		zlog.Error().Err(err).Msg("failed to declare night chat exchange")
		return  err
	}
	zlog.Info().Msg("exchanges declared")

	if err := cs.declareChatQueue(); err != nil {
		zlog.Error().Err(err).Msg("failed to declare chat queue")
	}
	zlog.Info().Msg("queue declared")

	if err := cs.bindDayChatWithQueue(); err != nil {
		zlog.Error().Err(err).Msg("failed to bind day chat with queue")
	}
	if err := cs.bindNightChatWithQueue(); err != nil {
		zlog.Error().Err(err).Msg("failed to bind night chat with queue")
	}
	zlog.Info().Msg("queue binded with exchanges")

	zlog.Info().Str("session id", cs.sessionId).Msg("chat for session started")
	
	return nil
}

func (cs *ChatServer) DeleteChat() error {
	errPref := zlog.Error().Str("session id", cs.sessionId).Str("queue name", cs.queue.Name)

	if _, err := cs.ch.QueueDelete(cs.queue.Name, false, false, false);
		err != nil {
		errPref.Err(err).Msg("failed to delete queue")
		return err
	}
	if err := cs.ch.ExchangeDelete(cs.GetSessionChatName(DAY_CHAT), false, false);
	   err != nil {
		errPref.Err(err).Msg("failed to delete day chat")
		return err
	}
	if err := cs.ch.ExchangeDelete(cs.GetSessionChatName(NIGHT_CHAT), false, false);
	   err != nil {
		errPref.Err(err).Msg("failed to delete night chat")
		return err
	}

	if err := cs.ch.Close(); err != nil {
		errPref.Err(err).Msg("failed to close rabbitmq channel")
		return err
	}
	if err := cs.conn.Close(); err != nil {
		errPref.Err(err).Msg("failed to close rabbitmq connection")
	}

	return nil
}

func (cs *ChatServer) declareDayChatExchange() error {
	return cs.declareExchange(cs.GetSessionChatName(DAY_CHAT))
}

func (cs *ChatServer) declareNightChatExchange() error {
	return cs.declareExchange(cs.GetSessionChatName(NIGHT_CHAT))
}

func (cs *ChatServer) declareExchange(name string) error {
	return cs.ch.ExchangeDeclare(
		name,     // exchange
		"fanout", // kind (send without binding keys; do not need them)
		false,    // durable (do not store sent and received messages after session end)
		false,    // autoDelete
		false,    // internal
		false,    // noWait (sync wait declaration)
		nil,      // args
	)

}

func (cs *ChatServer) declareChatQueue() error {
	queueName := cs.GetChatQueueName()
	queue, err := cs.ch.QueueDeclare(
		queueName,    // name
		false,        // durable (do not store messages after session end)
		false,        // autoDelete
		true,         // exclusive (chat will be used by only client)
		false,        // noWait (sync wait declaration)
		nil,		  // args
	)
	if err != nil {
		return err
	}

	cs.queue = queue
	return nil
}

func (cs *ChatServer) bindDayChatWithQueue() error {
	return cs.bindQueueWithExchange(cs.GetSessionChatName(DAY_CHAT))
}

func (cs *ChatServer) bindNightChatWithQueue() error {
	return cs.bindQueueWithExchange(cs.GetSessionChatName(NIGHT_CHAT))
}

func (cs *ChatServer) bindQueueWithExchange(exchangeName string) error {
	return cs.ch.QueueBind(
		cs.queue.Name, // name
		"",            // routing key (fanout ignores routing key)
		exchangeName,  // exchange
		false,         // noWait (sync wait declaration)
		nil,           // args
	)
}
