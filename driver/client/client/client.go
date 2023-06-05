package client

import (
	"context"
	mafiapb "driver/server/proto"
	"driver/support"
	"sync"

	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type client struct {
    username string
    userId string

    alivePlayers []string

    grpcClient mafiapb.MafiaDriverClient
    stream mafiapb.MafiaDriver_DoActionClient

    actionProducer actionProducerFunctor

    waitStartSessionMessage sync.WaitGroup
    waitUserIdAssignment sync.WaitGroup
    waitInteractorStartMsgs sync.WaitGroup
    waitAllUsersConnected sync.WaitGroup

    waitActionResponse chan interface{}
}

type actionProducerFunctor func (*client) error

func NewClient(mode, username string, conn *grpc.ClientConn) (*client, error) {

    if err := support.InitClientLogger(username);  err != nil {
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

    return &client{
        username: username,
        alivePlayers: make([]string, 0),
        grpcClient: mafiapb.NewMafiaDriverClient(conn),
        actionProducer: actionProducer,
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

func (c *client) SetAlivePlayers(nicknames []string) *client {
    c.alivePlayers = nicknames
    return c
}
