package client

import (
	"chat"
	"common"
	"context"
	mafiapb "driver/server/proto"
	"sync"
	"time"

	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type IClient interface {
    OpenStream(ctx context.Context) error
    StartPlaying() error
}

type client struct {
    username string
    userId string
    sessionId string
    auto bool
    spirit bool

    alivePlayers []string

    grpcClient mafiapb.MafiaDriverClient
    stream mafiapb.MafiaDriver_DoActionClient

    actionProducer actionProducerFunctor

    lastSessionTime time.Time
    rabbitmqConnParams *common.RabbitmqConnectionParams
    chat *chat.ChatServer

    curChat string
    chatsMsgs map[string][]*chat.ChatMessage
    msgsAccess sync.Mutex
    newMsgs chan interface{}

    waitStartSessionMessage sync.WaitGroup
    waitUserIdAssignment sync.WaitGroup
    waitInteractorStartMsgs sync.WaitGroup
    waitAllUsersConnected sync.WaitGroup

    waitActionResponse chan interface{}
}

type actionProducerFunctor func (*client) error

func NewManualClient(
    mode string,
    conn *grpc.ClientConn,
    rabbitmqConnParams *common.RabbitmqConnectionParams) (IClient, error) {
    return newClient(mode, "", conn, rabbitmqConnParams)
}

func NewAutoClient(
    mode, username string,
    conn *grpc.ClientConn,
    rabbitmqConnParams *common.RabbitmqConnectionParams) (IClient, error) {
    return newClient(mode, username, conn, rabbitmqConnParams)
}

func newClient(
    mode, username string,
    conn *grpc.ClientConn,
    rabbitmqConnParams *common.RabbitmqConnectionParams) (*client, error) {

    if err := common.InitClientLogger(username, mode == "auto");  err != nil {
        return nil, err
    }

    var actionProducer actionProducerFunctor
    switch mode {
    case "manual":
        actionProducer = manualActionProducer
    case "auto":
        actionProducer = automaticActionProducer
    default:
        panic("unknown client type")
    }

    chatsMsgs := make(map[string][]*chat.ChatMessage)
    chatsMsgs[chat.DAY_CHAT] = make([]*chat.ChatMessage, 0)
    chatsMsgs[chat.NIGHT_CHAT] = make([]*chat.ChatMessage, 0)

    return &client{
        username: username,
        auto: mode == "auto",
        spirit: false,
        alivePlayers: make([]string, 0),
        grpcClient: mafiapb.NewMafiaDriverClient(conn),
        actionProducer: actionProducer,
        lastSessionTime: time.Now(),
        rabbitmqConnParams: rabbitmqConnParams,
        curChat: chat.DAY_CHAT,
        chatsMsgs: chatsMsgs,
        newMsgs: make(chan interface{}, 1),
        waitActionResponse: make(chan interface{}, 1),
    }, nil
}

// Open stream with grpc server
func (c *client) OpenStream(ctx context.Context) error {
    stream, err := c.grpcClient.DoAction(ctx)
    if err != nil {
        return err
    }
    c.stream = stream
    return nil
}

// Send StartSession message to server
// and start playing
// Blocks until any error occured
func (c *client) StartPlaying() error {
    errorChan := make(chan error)

    c.waitStartSessionMessage.Add(1)
    c.waitUserIdAssignment.Add(1)
    c.waitInteractorStartMsgs.Add(1)
    c.waitAllUsersConnected.Add(1)

    go func() {
        if err := c.actionProducer(c); err != nil {
            zlog.Error().Err(err).Msg("failed when producing actions")
            errorChan <- err
        }
    }()

    go func() {
        if err := actionConsumer(c); err != nil {
            zlog.Error().Err(err).Msg("failed when consuming actions")
            errorChan <- err
        }
    }()

    return <-errorChan
}

func (c *client) GetAlivePlayers() *[]string {
    return &c.alivePlayers
}

func (c *client) GetChat() *chat.ChatServer {
    return c.chat
}

func (c *client) SetAlivePlayers(nicknames []string) *client {
    c.alivePlayers = nicknames
    return c
}

func (c *client) SetSessionId(sessionId string) *client {
    c.sessionId = sessionId
    return c
}

func (c *client) SetCurChat(chat string) *client {
    c.msgsAccess.Lock()
    defer c.msgsAccess.Unlock()

    c.curChat = chat

    return c
}

func (c *client) UpdateLastSessionTime() *client {
    c.msgsAccess.Lock()
    defer c.msgsAccess.Unlock()

    c.lastSessionTime = time.Now()

    return c
}

func (c *client) StartChat() error {
    if c.auto {
        return nil
    }

    chat := chat.NewChatServer(c.username, c.sessionId, c.rabbitmqConnParams)

    if err := chat.StartChat(); err != nil {
        zlog.Error().Err(err).Str("username", c.username).Msg("failed to start chat")
        return err
    }
    c.chat = chat
    zlog.Info().Str("username", c.username).Msg("client started chat")

    return nil
}
