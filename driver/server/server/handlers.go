package server

import (
	"driver/server/game"
	mafiapb "driver/server/proto"
	"fmt"
	"strconv"

	zlog "github.com/rs/zerolog/log"
)

func (s *Server) nextPlayerId() string {
	return strconv.Itoa(len(s.playerById))
}

func (s *Server) addPlayer(p *player, nickname string, session *game.GameSession) (bool, bool) {
    if _, has := s.nicknames[nickname]; has {
        return true, false
    }
    started, alreadyExists := session.AddPlayer(p.id, nickname)
    if alreadyExists {
        return true, false
    }
	s.playerById[p.id] = p
    s.nicknames[nickname] = struct{}{}
    s.idByNickname[session.GetPlayerNickname(p.id)] = p.id
	s.sessionPlayers[session] = append(s.sessionPlayers[session], p)
	s.sessionByUserId[p.id] = session

    return false, started
}

func (s *Server) removeGame(session *game.GameSession) {
    nicknames := session.GetAllPlayerNicknames()
    for _, nickname := range nicknames {
        delete(s.nicknames, nickname)
        id := s.idByNickname[nickname]
        delete(s.playerById, id)
        delete(s.sessionByUserId, id)
        delete(s.idByNickname, nickname)
    }
    delete(s.sessionPlayers, session)
}

// Returns true if game started
func (s *Server) handleStartSession(req *mafiapb.Action_StartSession, stream mafiapb.MafiaDriver_DoActionServer) error {
	username := req.GetNickname()
	logPref := zlog.Info().Str("username", username)

	logPref.Msg("accepted start session request")

	userSession := s.findOrCreatedNotStartedSession()
	logPref.Int("session id", userSession.Id).Msg("user added to session. waiting for game start")

	playerId := s.nextPlayerId()
	logPref.Str("player id", playerId).Msg("user assigned an id")

	newPlayer := &player{
		id: playerId,
		stream: stream,
	}
	alreadyExists, started := s.addPlayer(newPlayer, username, userSession)
    if alreadyExists {
        assignUserMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_ASSIGN_USER_ID,
            ActionResult: &mafiapb.ActionResponse_AssignUserId_{
                AssignUserId: &mafiapb.ActionResponse_AssignUserId{
                    Result: &mafiapb.ActionResponse_ActionResult{
                        Result: &mafiapb.ActionResponse_ActionResult_Error{
                            Error: "User with the same nickname already registered. Please change yours",
                        },
                    },
                },
            },
        }
        logPref.Str("message", assignUserMessage.String()).Msg("sending to user")
        if err := stream.Send(assignUserMessage); err != nil {
            return err
        }
        return nil
    }

	logPref.Bool("game started", started).Msg("game state")

	assignUserMessage := &mafiapb.ActionResponse{
		Type: mafiapb.ActionResponse_ASSIGN_USER_ID,
		ActionResult: &mafiapb.ActionResponse_AssignUserId_{
			AssignUserId: &mafiapb.ActionResponse_AssignUserId{
                Result: &mafiapb.ActionResponse_ActionResult{
                    Result: &mafiapb.ActionResponse_ActionResult_Success{
                        Success: playerId,
                    },
                },
			},
		},
	}
	logPref.Str("message", assignUserMessage.String()).Msg("sending to user")
	if err := s.sendMessageToPlayer(newPlayer, assignUserMessage); err != nil {
		return err
	}

	if started {
        userSession.StartGame()
        genGameStartMessageCb := func (p *player, session *game.GameSession)  *mafiapb.ActionResponse {
            playerRole := session.GetPlayerRoleString(p.id)
            return &mafiapb.ActionResponse{
                Type: mafiapb.ActionResponse_START_GAME,
                ActionResult: &mafiapb.ActionResponse_StartGame_{
                    StartGame: &mafiapb.ActionResponse_StartGame{
                        StartGame: fmt.Sprintf(
                            "\nGame started, your role is [%s], you are connected to session [%d]\n" +
                            "First day is introductory:\n" +
                            "- nobody will be killed\n" +
                            "- your vote won't be counted\n" +
                            "You can familiarize yourself with the interface",
                            playerRole, userSession.Id),
                        Nicknames: userSession.GetPlayerNicknames(),
                    },
                },
		    }
        }
        if err := s.sendMessageToGameSessionWithCb(userSession, genGameStartMessageCb); err != nil {
            return err
        }
		zlog.Info().Msg("game started")
	}

    return nil
}

func (s *Server) handlePlayerRoleRequest(p *player) error {
    logPref := zlog.Info().Str("id", p.id)
    logPref.Msg("accepted player role request")

    session := s.sessionByUserId[p.id]
    playerRole := session.GetPlayerRole(p.id)
    playerRoleMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_PLAYER_ROLE,
        ActionResult: &mafiapb.ActionResponse_Role{
            Role: &mafiapb.ActionResponse_PlayerRole{
                Role: mafiapb.ActionResponse_EPlayerRole(playerRole),
            },
        },
    }
    return s.sendMessageToPlayer(p, playerRoleMessage)
}

func (s *Server) handleGameStateRequest(p *player) error {
    logPref := zlog.Info().Str("id", p.id)
    logPref.Msg("accepted game state request")

    session := s.sessionByUserId[p.id]
    currentDay := session.GetDay()
    alivePlayers := session.GetAlivePlayersCount()
    gameStateMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_GAME_STATE,
        ActionResult: &mafiapb.ActionResponse_GameState_{
            GameState: &mafiapb.ActionResponse_GameState{
                CurrentDay: int32(currentDay),
                AlivePlayers: int32(alivePlayers),
            },
        },
    }
    return s.sendMessageToPlayer(p, gameStateMessage)
}

func (s *Server) handlePlayerNicksRequest(p *player) error {
    logPref := zlog.Info().Str("id", p.id)
    logPref.Msg("accepted player nicks request")

    session := s.sessionByUserId[p.id]
    nicknames := session.GetPlayerNicknames()
    playerNicksMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_PLAYER_NICKS,
        ActionResult: &mafiapb.ActionResponse_PlayerNicks_{
            PlayerNicks: &mafiapb.ActionResponse_PlayerNicks{
                Nicks: nicknames,
            },
        },
    }

    return s.sendMessageToPlayer(p, playerNicksMessage)
}

func (s *Server) handleInterruptGameRequest(p *player) error {
    logPref := zlog.Info().Str("id", p.id)
    logPref.Msg("accepted interrupt game request")

    session := s.sessionByUserId[p.id]
    nickname := session.GetPlayerNickname(p.id)
    gameEndMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_END_GAME,
        ActionResult: &mafiapb.ActionResponse_EndGame_{
            EndGame: &mafiapb.ActionResponse_EndGame{
                GameResult: fmt.Sprintf(
                    "game was interrupted by '%s'", nickname),
            },
        },
    }

    return s.sendMessageToGameSession(session, gameEndMessage)
}

func (s *Server) handleVoteRequest(p *player, suspectNick string) error {
    logPref := zlog.Info().Str("id", p.id).Str("suspect", suspectNick)
    logPref.Msg("accepted vote request")

    session := s.sessionByUserId[p.id]
    nightStarted, wrongVoteErr := session.Vote(p.id, suspectNick)
    if wrongVoteErr != nil {
        zlog.Error().Err(wrongVoteErr).Msg("bad vote")

        wrongVoteMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_VOTE_RESULT,
            ActionResult: &mafiapb.ActionResponse_VoteResult_{
                VoteResult: &mafiapb.ActionResponse_VoteResult{
                    Result: &mafiapb.ActionResponse_ActionResult{
                        Result: &mafiapb.ActionResponse_ActionResult_Error{
                            Error: wrongVoteErr.Error(),
                        },
                    },
                },
            },
        }
        if err := s.sendMessageToPlayer(p, wrongVoteMessage); err != nil {
            return err
        }
        return nil
    }

    voteResultMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_VOTE_RESULT,
        ActionResult: &mafiapb.ActionResponse_VoteResult_{
            VoteResult: &mafiapb.ActionResponse_VoteResult{
                Result: &mafiapb.ActionResponse_ActionResult{
                    Result: &mafiapb.ActionResponse_ActionResult_Success{
                        Success: "vote accepted. Wait until all players make their votes",
                    },
                },
            },
        },
    } 
    if err := s.sendMessageToPlayer(p, voteResultMessage); err != nil {
        return err
    }

    if nightStarted {
        return s.handleStartNightEvent(session)
    }

    return nil
}

func (s *Server) handleStartNightEvent(session *game.GameSession) error {
    nightInfo := session.GetNightInfo()
    zlog.Info().Interface("game state", nightInfo).Msg("Night started")

    mafia := s.playerById[nightInfo.MafiaId]
    mafiaNightMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_NIGHT_STARTED,
        ActionResult: &mafiapb.ActionResponse_NightStarted_{
            NightStarted: &mafiapb.ActionResponse_NightStarted{
                UserMsg: "night started, your are mafia. Choose a player to kill",
                Role: mafiapb.ActionResponse_MAFIA,
            },
        },
    }
    zlog.Info().Str("msg", mafiaNightMessage.String()).Msg("sending message to mafia")
    if err := s.sendMessageToPlayer(mafia, mafiaNightMessage); err != nil {
        return err
    }

    if comissarId := nightInfo.ComissarId; len(comissarId) > 0 {
        comissar := s.playerById[comissarId]

        comissarNightMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_NIGHT_STARTED,
            ActionResult: &mafiapb.ActionResponse_NightStarted_{
                NightStarted: &mafiapb.ActionResponse_NightStarted{
                    UserMsg: "night started, your are commisar. Choose a player to investigate",
                    Role: mafiapb.ActionResponse_COMMISAR,
                },
            },
        }
        zlog.Info().Str("msg", comissarNightMessage.String()).Msg("sending message to comissar")
        if err := s.sendMessageToPlayer(comissar, comissarNightMessage); err != nil {
            return err
        }
    }

    for _, civilianId := range nightInfo.CivilianIds {
        civilian := s.playerById[civilianId]

        civilianNightMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_NIGHT_STARTED,
            ActionResult: &mafiapb.ActionResponse_NightStarted_{
                NightStarted: &mafiapb.ActionResponse_NightStarted{
                    UserMsg: "night started, you are civilian. Wait until mafia and commisar made their decisions",
                    Role: mafiapb.ActionResponse_CIVILIAN,
                },
            },
        }
        zlog.Info().Str("msg", civilianNightMessage.String()).Msg("sending message to civilian")
        if err := s.sendMessageToPlayer(civilian, civilianNightMessage); err != nil {
            return err
        }
    }

    return nil
}

func (s *Server) handleInvestigationResultEvent(commissar *player, publishResult bool) error {
    logPref := zlog.Info().Str("commissar id", commissar.id)
    logPref.Msg("accepted investigation result request")

    session := s.sessionByUserId[commissar.id]
    if session.CommissarFoundMafia(publishResult) {
        s.handleStartDayEvent(session)
    }

    return nil
}


func (s *Server) handleMafiaKillPlayerRequest(mafiaId string, p *player) error {
    mafia := s.playerById[mafiaId]
    session := s.sessionByUserId[mafiaId]

    if p == nil {
        wrongVoteMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_MAFIA_KILL_RESULT,
            ActionResult: &mafiapb.ActionResponse_MafiaKillResult_{
                MafiaKillResult: &mafiapb.ActionResponse_MafiaKillResult{
                    Result: &mafiapb.ActionResponse_ActionResult{
                        Result: &mafiapb.ActionResponse_ActionResult_Error{
                            Error: "no such player",
                        },
                    },
                },
            },
        }

        if err := s.sendMessageToPlayer(mafia, wrongVoteMessage); err != nil {
            return err
        }
        return nil
    }

    startDay, wrongVoteErr := session.AcceptMafiaVote(mafiaId, p.id)
    if wrongVoteErr != nil {
        wrongVoteMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_MAFIA_KILL_RESULT,
            ActionResult: &mafiapb.ActionResponse_MafiaKillResult_{
                MafiaKillResult: &mafiapb.ActionResponse_MafiaKillResult{
                    Result: &mafiapb.ActionResponse_ActionResult{
                        Result: &mafiapb.ActionResponse_ActionResult_Error{
                            Error: wrongVoteErr.Error(),
                        },
                    },
                },
            },
        }

        if err := s.sendMessageToPlayer(mafia, wrongVoteMessage); err != nil {
            return err
        }
        return nil
    }

    logPref := zlog.Info().Str("mafia id", mafiaId).Str("id", p.id)
    logPref.Msg("accepted mafia kill player request")

    mafiaKillResultMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_MAFIA_KILL_RESULT,
        ActionResult: &mafiapb.ActionResponse_MafiaKillResult_{
            MafiaKillResult: &mafiapb.ActionResponse_MafiaKillResult{
                Result: &mafiapb.ActionResponse_ActionResult{
                    Result: &mafiapb.ActionResponse_ActionResult_Success{
                        Success: "vote accepted",
                    },
                },
            },
        },
    }

    if err := s.sendMessageToPlayer(mafia, mafiaKillResultMessage); err != nil {
        return err
    }

    if startDay {
        return s.handleStartDayEvent(session)
    }

    return nil
}

func (s *Server) handleInvestigateMafiaRequest(commissarId string, p *player) error {
    commissar := s.playerById[commissarId]

    if p == nil {
        wrongVoteMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_COMMISAR_INVESTIGATION_RESULT,
            ActionResult: &mafiapb.ActionResponse_ComissareInvestigationResult{
                ComissareInvestigationResult: &mafiapb.ActionResponse_CommisarInvestigationResult{
                    Result: &mafiapb.ActionResponse_ActionResult{
                        Result: &mafiapb.ActionResponse_ActionResult_Error{
                            Error: "no such player",
                        },
                    },
                },
            },
        }

        if err := s.sendMessageToPlayer(commissar, wrongVoteMessage); err != nil {
            return err
        }
        return nil
    }

    session := s.sessionByUserId[p.id]
    wrongVoteErr := session.AcceptCommissarVote(commissarId, p.id)
    if wrongVoteErr != nil {
        wrongVoteMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_COMMISAR_INVESTIGATION_RESULT,
            ActionResult: &mafiapb.ActionResponse_ComissareInvestigationResult{
                ComissareInvestigationResult: &mafiapb.ActionResponse_CommisarInvestigationResult{
                    Result: &mafiapb.ActionResponse_ActionResult{
                        Result: &mafiapb.ActionResponse_ActionResult_Error{
                            Error: wrongVoteErr.Error(),
                        },
                    },
                },
            },
        }

        if err := s.sendMessageToPlayer(commissar, wrongVoteMessage); err != nil {
            return err
        }
        return nil
    }

    logPref := zlog.Info().Str("commissar id", commissarId).Str("id", p.id)
    logPref.Msg("accepted investigate mafia request")

    var result string
    if session.IsMafia(p.id) {
        result = "mafia"
    } else {
        result = "not mafia"
    }

    investigationResultMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_COMMISAR_INVESTIGATION_RESULT,
        ActionResult: &mafiapb.ActionResponse_ComissareInvestigationResult{
            ComissareInvestigationResult: &mafiapb.ActionResponse_CommisarInvestigationResult{
                Result: &mafiapb.ActionResponse_ActionResult{
                    Result: &mafiapb.ActionResponse_ActionResult_Success{
                        Success: result,
                    },
                },
                MafiaNickname: session.GetPlayerNickname(p.id),
            },
        },
    }

    return s.sendMessageToPlayer(commissar, investigationResultMessage)
}

func (s *Server) handleStartDayEvent(session *game.GameSession) error {
    morningSummary := session.GetMorningSummary()

    roundResultMessage := &mafiapb.ActionResponse{
        Type: mafiapb.ActionResponse_ROUND_RESULT,
        ActionResult: &mafiapb.ActionResponse_Result{
            Result: &mafiapb.ActionResponse_RoundResult{
                UserMsg: fmt.Sprintf(
                    "In the night something has happened:\n" +
                    "- Was killed by voting decision: %s\n" +
                    "- Was killed by mafia: %s\n" +
                    "- Commissar investigation result: %s\n",
                    morningSummary.KilledPlayerNickname,
                    morningSummary.KilledByMafiaPlayerNickname,
                    morningSummary.CommissarInvestigationResult),
            },
        },
    }

    if err := s.sendMessageToGameSession(session, roundResultMessage); err != nil {
        return err
    }

    gameMorningStatus := session.StartDay()
    whoWon := func() string {
        switch gameMorningStatus {
        case game.MAFIAN_WON:
            return "mafia won"
        case game.CIVILIAN_WON:
            return "civilian won"
        default:
            return "nobody"
        }
    }()

    switch gameMorningStatus {
    case game.NOT_FINISHED:
        dayStartedMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_DAY_STARTED,
            ActionResult: &mafiapb.ActionResponse_DayStarted_{
                    DayStarted: &mafiapb.ActionResponse_DayStarted{
                        UserMsg: fmt.Sprintf(
                            "Day %d passed\n" +
                            "Procceed to vote to find the mafia!",
                        session.GetDay() - 1),
                    Nicknames: session.GetPlayerNicknames(),
                },
            },
        }

        if err := s.sendMessageToGameSession(session, dayStartedMessage); err != nil {
            return err
        }
    default:
        endGameMessage := &mafiapb.ActionResponse{
            Type: mafiapb.ActionResponse_END_GAME,
            ActionResult: &mafiapb.ActionResponse_EndGame_{
                EndGame: &mafiapb.ActionResponse_EndGame{
                    GameResult: whoWon,
                },
            },
        }

        if err := s.sendMessageToGameSession(session, endGameMessage); err != nil {
            return err
        }

        s.removeGame(session)
    }

    return nil
}
