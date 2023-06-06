package client

import (
	mafiapb "driver/server/proto"
	"math/rand"

	zlog "github.com/rs/zerolog/log"
)

func automaticActionProducer(c *client) error {
    stream := c.stream

    // send start session request
    startSessionReq := &mafiapb.Action{
        Type: mafiapb.Action_START_SESSION,
        Action: &mafiapb.Action_StartSession_{
            StartSession: &mafiapb.Action_StartSession{
                Nickname: c.username,
            },
        },
    }
    if err := stream.Send(startSessionReq); err != nil {
        zlog.Error().Err(err).Str("msg", startSessionReq.String()).Msg("failed to start session")
        return err
    }

    c.waitStartSessionMessage.Done()

    c.waitUserIdAssignment.Wait()
    zlog.Info().Str("id", c.userId).Str("name", c.username).Msg("connected to game")
    c.waitInteractorStartMsgs.Done()
    c.waitAllUsersConnected.Wait()

    errChan := make(chan error)
    nextActionChan := make(chan interface{}, 1)
    firstAction := true
    go func() {
        for {
            if !firstAction {
                <-nextActionChan
            }
            firstAction = false

            if c.spirit {
                zlog.Info().Str("name", c.username).Msg("bot is spirit, wait do nothing")
                return
            }

            nickname := c.chooseRandomPlayer()
            zlog.Info().Str("nickname", nickname).Str("name", c.username).Msg("bot votes")

            if err := c.sendVote(nickname); err != nil {
                errChan <- err
                return
            }
        }
    }()

    for {
        zlog.Info().Str("name", c.username).Msg("waits for next action")
        <-c.waitActionResponse
        nextActionChan <- struct{}{}
    }
}

func (c *client) chooseRandomPlayer() string {
    for {
        nicknames := *c.GetAlivePlayers()
        nickname := nicknames[rand.Intn(len(nicknames))]
        if nickname != c.username {
            return nickname
        }
    }
}
