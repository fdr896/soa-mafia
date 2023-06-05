package client

import (
	mafiapb "driver/server/proto"

	zlog "github.com/rs/zerolog/log"
)

func (c *client) sendRoleReq() error {
    roleReq := &mafiapb.Action{
        Type: mafiapb.Action_PLAYER_ROLE,
        Action: &mafiapb.Action_PlayerRole_{
            PlayerRole: &mafiapb.Action_PlayerRole{
                UserId: c.userId,
            },
        },
    }

    if err := c.stream.Send(roleReq); err != nil {
        zlog.Error().Err(err).Str("msg", roleReq.String()).Msg("failed to send role request")
        return err
    }

    return nil
}

func (c *client) sendStateReq() error {
    stateReq := &mafiapb.Action{
        Type: mafiapb.Action_GAME_STATE,
        Action: &mafiapb.Action_GameState_{
            GameState: &mafiapb.Action_GameState{
                UserId: c.userId,
            },
        },
    }

    if err := c.stream.Send(stateReq); err != nil {
        zlog.Error().Err(err).Str("msg", stateReq.String()).Msg("failed to send state request")
        return err
    }

    return nil
}

func (c *client) sendNicksReq() error {
    nicksReq := &mafiapb.Action{
        Type: mafiapb.Action_PLAYER_NICKS,
        Action: &mafiapb.Action_PlayerNicks{
            PlayerNicks: &mafiapb.Action_PlayersNicks{
                UserId: c.userId,
            },
        },
    }

    if err := c.stream.Send(nicksReq); err != nil {
        zlog.Error().Err(err).Str("msg", nicksReq.String()).Msg("failed to send nicks request")
        return err
    }

    return nil
}

func (c *client) sendInterruptReq() error {
    interruptReq := &mafiapb.Action{
        Type: mafiapb.Action_INTERRUPT_GAME,
        Action: &mafiapb.Action_InterruptGame_{
            InterruptGame: &mafiapb.Action_InterruptGame{
                UserId: c.userId,
            },
        },
    }

    if err := c.stream.Send(interruptReq); err != nil {
        zlog.Error().Err(err).Str("msg", interruptReq.String()).Msg("failed to send interrupt request")
        return err
    }

    return nil
}

func (c *client) sendVote(suspectNick string) error {
    voteReq := &mafiapb.Action{
        Type: mafiapb.Action_VOTE,
        Action: &mafiapb.Action_Vote_{
            Vote: &mafiapb.Action_Vote{
                UserId: c.userId,
                MafiaUsername: suspectNick,
            },
        },
    }

    if err := c.stream.Send(voteReq); err != nil {
        zlog.Error().Err(err).Str("msg", voteReq.String()).Msg("failed to vote request")
        return err
    }

    return nil
}