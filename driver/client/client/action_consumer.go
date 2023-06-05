package client

import (
	"fmt"
	"io"
	"os"
	"strings"

	mafiapb "driver/server/proto"

	zlog "github.com/rs/zerolog/log"
)

func actionConsumer(c *client) error {
    stream := c.stream
    ctx := stream.Context()

    c.waitStartSessionMessage.Wait()

    for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

        action, err := stream.Recv()
		if err == io.EOF {
			zlog.Info().Msg("stream ended")
			return nil
		}
		if err != nil {
			zlog.Error().Err(err).Msg("failed to receive message from stream")
            return err
		}

        zlog.Info().Str("msg", action.String()).Msg("received msg")
        if err := handleReceivedActionResp(c, action); err != nil {
            zlog.Error().Err(err).Str("action", action.String()).Msg("failed to handle received action")
            return err
        }
    }
}

func handleReceivedActionResp(c *client, action *mafiapb.ActionResponse) error {
    switch action.GetType() {
    case mafiapb.ActionResponse_ASSIGN_USER_ID:
        assignUserId := action.GetAssignUserId()
        result := assignUserId.GetResult()
        if userId := result.GetSuccess(); len(userId) > 0 {
            c.userId = userId
        } else if errMsg := result.GetError(); len(errMsg) > 0 {
            fmt.Println(errMsg)
            os.Exit(0)
        } else {
            panic("inconsistent msg")
        }
        c.waitUserIdAssignment.Done()
    case mafiapb.ActionResponse_PLAYER_ROLE:
        playerRole := action.GetRole()
        fmt.Printf("Your role is: %s\n", playerRole.GetRole())
        c.waitActionResponse <- struct{}{}
    case mafiapb.ActionResponse_GAME_STATE:
        gameState := action.GetGameState()
        fmt.Printf("Alive players: %d\nCurrently is %s\n",
            gameState.GetAlivePlayers(),
            mafiapb.ActionResponse_ETimeOfDay_name[int32(gameState.GetTimeOfDay())])
        c.waitActionResponse <- struct{}{}
    case mafiapb.ActionResponse_PLAYER_NICKS:
        playerNicks := action.GetPlayerNicks()
        fmt.Printf("Alive player nicks: %+q\n", playerNicks.GetNicks())
        c.waitActionResponse <- struct{}{}
    case mafiapb.ActionResponse_VOTE_RESULT:
        voteResult := action.GetVoteResult()
        result := voteResult.GetResult()
        if successMsg := result.GetSuccess(); len(successMsg) > 0 {
            fmt.Printf("Vote result: %s\n", successMsg)
        } else if errMsg := result.GetError(); len(errMsg) > 0 {
            fmt.Printf("Vote result: %s\n", errMsg)
            fmt.Println("Vote again")
            c.waitActionResponse <- struct{}{}
        } else {
            panic("inconsistent msg")
        }
    case mafiapb.ActionResponse_MAFIA_KILL_RESULT:
        mafiaKillResult := action.GetMafiaKillResult()
        result := mafiaKillResult.GetResult()
        if success := result.GetSuccess(); len(success) > 0 {
            fmt.Println("Player will be killed")
        } else if errMsg := result.GetError(); len(errMsg) > 0 {
            fmt.Printf("Failed to kill the player: %s\n", errMsg)
            if err := c.acceptMafiaKillUsername(); err != nil {
                return err
            }
        } else {
            panic("inconsistent msg")
        }
    case mafiapb.ActionResponse_COMMISAR_INVESTIGATION_RESULT:
        comissarInvestigationResult := action.GetComissareInvestigationResult()
        result := comissarInvestigationResult.GetResult()
        if investigationResult := result.GetSuccess(); len(investigationResult) > 0 {
            fmt.Printf("Investigation result: %s\n", investigationResult)

            if investigationResult == "mafia" {
                if err := c.acceptComissarPublishResultDesire(comissarInvestigationResult.GetMafiaNickname()); err != nil {
                    return err
                }
            } else {
                investigationResultMessage := &mafiapb.Action{
                    Type: mafiapb.Action_INVESTIGATION_RESULT,
                    Action: &mafiapb.Action_InvestigationResult_{
                        InvestigationResult: &mafiapb.Action_InvestigationResult{
                            UserId: c.userId,
                            PublishResult: false,
                        },
                    },
                }

                if err := c.stream.Send(investigationResultMessage); err != nil {
                    return err
                }
            }
        } else if errMsg := result.GetError(); len(errMsg) > 0 {
            fmt.Printf("Failed to invesigate the player: %s\n", errMsg)
            if err := c.acceptComissarInvestigateUsername(); err != nil {
                return err
            }
        } else {
            panic("inconsistent msg")
        }
    case mafiapb.ActionResponse_START_GAME:
        c.waitInteractorStartMsgs.Wait()
        startGame := action.GetStartGame()
        fmt.Println(startGame.GetStartGame())
        c.waitAllUsersConnected.Done()
    case mafiapb.ActionResponse_END_GAME:
        endGame := action.GetEndGame()
        fmt.Println("\nGame finished:", endGame.GetGameResult())
        c.stream.CloseSend()
        os.Exit(0)
    case mafiapb.ActionResponse_DAY_STARTED:
        dayStarted := action.GetDayStarted()
        fmt.Printf("Morning started: %s\n", dayStarted.GetUserMsg())

        c.waitActionResponse <- struct{}{}
    case mafiapb.ActionResponse_NIGHT_STARTED:
        nightStarted := action.GetNightStarted()
        fmt.Printf("Night message: %s\n", nightStarted.GetUserMsg())

        switch nightStarted.GetRole() {
        case mafiapb.ActionResponse_MAFIA:
            if err := c.acceptMafiaKillUsername(); err != nil {
                return err
            }
        case mafiapb.ActionResponse_COMMISAR:
            if err := c.acceptComissarInvestigateUsername(); err != nil {
                return err
            }
        case mafiapb.ActionResponse_CIVILIAN:
        }
    case mafiapb.ActionResponse_ROUND_RESULT:
        roundResult := action.GetResult()
        fmt.Printf("Round result:\n%s\n", roundResult.GetUserMsg())
    }

    return nil
}

func (c *client) acceptMafiaKillUsername() error {
    var nickname string
    for {
        fmt.Print("Type nickname to kill: ")
        _, err := fmt.Scanln(&nickname)
        nickname = strings.TrimSpace(nickname)
        if err == nil && len(nickname) > 0 {
            break
        }
    }
    nickname = strings.Split(strings.TrimSpace(nickname), " ")[0]

    killPlayer := &mafiapb.Action{
        Type: mafiapb.Action_KILL_PLAYER_BY_MAFIA,
        Action: &mafiapb.Action_KillPlayerByMafia_{
            KillPlayerByMafia: &mafiapb.Action_KillPlayerByMafia{
                UserId: c.userId,
                PlayerUsername: nickname,
            },
        },
    }
    if err := c.stream.Send(killPlayer); err != nil {
        return err
    }

    return nil
}

func (c *client) acceptComissarInvestigateUsername() error {
    var nickname string
    for {
        fmt.Print("Type nickname to investigate: ")
        _, err := fmt.Scanln(&nickname)
        nickname = strings.TrimSpace(nickname)
        if err == nil && len(nickname) > 0 {
            break
        }
    }
    nickname = strings.Split(strings.TrimSpace(nickname), " ")[0]

    investigatePlayer := &mafiapb.Action{
        Type: mafiapb.Action_INVESTIGATE_MAFIA,
        Action: &mafiapb.Action_InvestiageMafia{
            InvestiageMafia: &mafiapb.Action_InvestigateMafia{
                UserId: c.userId,
                MafiaUsername: nickname,
            },
        },
    }
    if err := c.stream.Send(investigatePlayer); err != nil {
        return err
    }

    return nil
}

func (c *client) acceptComissarPublishResultDesire(mafiaNickname string) error {
    var desire string
    for {
        fmt.Println("\nType 'yes' or 'no' whether you want to publish the results: ")
        var yesOrNo string
        for {
            _, err := fmt.Scanln(&yesOrNo)
            yesOrNo = strings.TrimSpace(yesOrNo)
            if err == nil && len(yesOrNo) > 0 {
                break
            }
        }

        if yesOrNo != "yes" && yesOrNo != "no" {
            continue
        }

        desire = yesOrNo
        break
    }

    var publishResult bool
    switch desire {
    case "yes":
        publishResult = true
    case "no":
        publishResult = false
    default:
        panic("bits flipped!")
    }

    investigationResultMessage := &mafiapb.Action{
        Type: mafiapb.Action_INVESTIGATION_RESULT,
        Action: &mafiapb.Action_InvestigationResult_{
            InvestigationResult: &mafiapb.Action_InvestigationResult{
                UserId: c.userId,
                PublishResult: publishResult,
            },
        },
    }

    return c.stream.Send(investigationResultMessage)
}