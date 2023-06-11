package pdf

import (
	"common"
	"context"
	"encoding/json"
	"stat_manager/storage/database"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	zlog "github.com/rs/zerolog/log"
)

const (
	QUEUE_NAME = "tasks_queue"
)

type TaskManager struct {
	connParams *common.RabbitmqConnectionParams

	conn *amqp.Connection
	ch *amqp.Channel
	queue amqp.Queue
}

func NewTaskManager(connParams *common.RabbitmqConnectionParams) *TaskManager {
	return &TaskManager{
		connParams: connParams,
	}
}

func (tm *TaskManager) Start() error {
	connUrl := common.GetRabbitmqConnectionUrl(tm.connParams)
	zlog.Info().Str("url", connUrl).Msg("connecting to rabbitmq")

	conn, err := amqp.Dial(connUrl)
	if err != nil {
		return err
	}
	tm.conn = conn
	zlog.Info().Msg("connection establised")

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	tm.ch = ch
	zlog.Info().Msg("chan created")

	if err := tm.declareQueue(); err != nil {
		zlog.Error().Err(err).Msg("failed to declare queue")
	}
	zlog.Info().Msg("queue declared")

	return nil
}

func (tm *TaskManager) SubmitPdfGenTask(p *database.Player) error {
	pBytes, err := json.Marshal(*p)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	zlog.Debug().Interface("player", *p).Msg("submit player pdf gen task")
	return tm.ch.PublishWithContext(
		ctx,              // ctx (timeout after some time)
		"",               // exchange
		tm.queue.Name,    // routing key
		false,            // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body: pBytes,
		},     // msg
	)
}

func (tm *TaskManager) ReceivePdfGenTasks() (<-chan amqp.Delivery, error) {
	zlog.Info().Str("queue", tm.queue.Name).Msg("start consuming")

	tasksChan, err := tm.ch.Consume(
		tm.queue.Name, // queue
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

	return tasksChan, nil
}

func (tm *TaskManager) declareQueue() error {
	queue, err := tm.ch.QueueDeclare(
		QUEUE_NAME,   // name
		false,        // durable (do not store messages after session end)
		false,        // autoDelete
		true,         // exclusive (chat will be used by only client)
		false,        // noWait (sync wait declaration)
		nil,		  // args
	)
	if err != nil {
		return err
	}

	tm.queue = queue
	return nil
}
