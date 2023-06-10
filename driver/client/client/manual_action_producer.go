package client

import (
	"driver/cli"
	mafiapb "driver/server/proto"
	"fmt"
	"strconv"

	zlog "github.com/rs/zerolog/log"
)

func manualActionProducer(c *client) error {
    cmdCli := cli.NewCliInteractor()

    c.username = cmdCli.Start()

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
    cmdCli.Print(fmt.Sprintf("Your user id is [%s]", c.userId))

    cmdCli.Print("Waiting until all users are connected...")
    c.waitInteractorStartMsgs.Done()

    c.waitAllUsersConnected.Wait()

    cmdCli.PrintRules()

    errChan := make(chan error)

    go c.StartReadingChat(errChan)

    go func() {
        for cmd := range cmdCli.Commands() {
            switch cmd.GetType() {
            case cli.CMD_ROLE:
                if err := c.sendRoleReq(); err != nil {
                    errChan <- err
                    return
                }
            case cli.CMD_STATE:
                if err := c.sendStateReq(); err != nil {
                    errChan <- err
                    return
                }
            case cli.CMD_NICKS:
                if err := c.sendNicksReq(); err != nil {
                    errChan <- err
                    return
                }
            case cli.CMD_VOTE:
                if c.handleSpiritClientActionResult() {
                    continue
                }

                if err := c.sendVote(cmd.GetArg(0)); err != nil {
                    errChan <- err
                    return
                }
            case cli.CMD_EXIT:
                if err := c.sendInterruptReq(); err != nil {
                    errChan <- err
                    return
                }
            case cli.CMD_RULES:
                cmdCli.PrintRules()
                c.waitActionResponse <- struct{}{}
            case cli.CMD_READ:
                if c.handleSpiritClientActionResult() {
                    continue
                }

                if err := c.ReadChatSession(); err != nil {
                    errChan <- err
                    return
                }
                c.waitActionResponse <- struct{}{}
            case cli.CMD_READ_ALL:
                if c.handleSpiritClientActionResult() {
                    continue
                }

                if err := c.ReadChatAll(); err != nil {
                    errChan <- err
                    return
                }
                c.waitActionResponse <- struct{}{}
            case cli.CMD_READ_LAST_N:
                if c.handleSpiritClientActionResult() {
                    continue
                }

                n, _ := strconv.Atoi(cmd.GetArg(0))
                if err := c.ReadChatLastN(n); err != nil {
                    errChan <- err
                    return
                }
                c.waitActionResponse <- struct{}{}
            case cli.CMD_WRITE:
                if c.handleSpiritClientActionResult() {
                    continue
                }

                if err := c.WriteChat(); err != nil {
                    errChan <- err
                    return
                }
                c.waitActionResponse <- struct{}{}
            }
        }
    }()

    cmdCli.StartPlaying(c.waitActionResponse, c.GetAlivePlayers())

    return <-errChan
}

func (c *client) handleSpiritClientActionResult() bool {
    if c.spirit {
        fmt.Println("Your are a spirit, chat and voting is unavailable")
        c.waitActionResponse <- struct{}{}

        return true
    } else {
        return false
    }
}
