package server

import (
	"driver/server/game"
	mafiapb "driver/server/proto"
	"io"
	"sync"

	zlog "github.com/rs/zerolog/log"
)

type Server struct {
	mafiapb.UnimplementedMafiaDriverServer

    gamePlayers int

	playerById map[string]*player
    nicknames map[string]interface{}
    idByNickname map[string]string
	sessions map[*game.GameSession]interface{}
	sessionByUserId map[string]*game.GameSession
	sessionPlayers map[*game.GameSession][]*player

	mutex sync.Mutex
}

type player struct {
	id string
	stream mafiapb.MafiaDriver_DoActionServer
}

func NewServer(gamePlayers int) *Server {
	return &Server{
        gamePlayers: gamePlayers,
		playerById: make(map[string]*player),
        nicknames: make(map[string]interface{}),
        idByNickname: make(map[string]string),
		sessions: make(map[*game.GameSession]interface{}),
		sessionByUserId: make(map[string]*game.GameSession),
		sessionPlayers: make(map[*game.GameSession][]*player),
	}
}

func (s *Server) DoAction(stream mafiapb.MafiaDriver_DoActionServer) error {
	errorChan := make(chan error)

	go s.listenFromPlayerStream(stream, errorChan)

	return <-errorChan
}

func (s *Server) listenFromPlayerStream(stream mafiapb.MafiaDriver_DoActionServer, errorChan chan error) {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			errorChan <- ctx.Err()
			return
		default:
		}

		action, err := stream.Recv()
		if err == io.EOF {
			zlog.Info().Msg("stream ended")
			errorChan <- nil
			return
		}
		if err != nil {
			zlog.Error().Err(err).Msg("failed to received message from stream")
			errorChan <- err
			return
		}

        if err := s.handleReceivedAction(stream, action); err != nil {
            zlog.Error().Err(err).Str("action", action.String()).Msg("failed to handle received action")
            errorChan <- err
            return
        }
	}
}

func (s *Server) handleReceivedAction(stream mafiapb.MafiaDriver_DoActionServer, action *mafiapb.Action) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    switch action.GetType() {
    case mafiapb.Action_START_SESSION:
        startSessionReq := action.GetStartSession()
        if err := s.handleStartSession(startSessionReq, stream); err != nil {
            return err
        }
    case mafiapb.Action_PLAYER_ROLE:
        getRoleReq := action.GetPlayerRole()
        player := s.playerById[getRoleReq.GetUserId()]
        if err := s.handlePlayerRoleRequest(player); err != nil {
            return err
        }
    case mafiapb.Action_GAME_STATE:
        getGameStateReq := action.GetGameState()
        player := s.playerById[getGameStateReq.GetUserId()]
        if err := s.handleGameStateRequest(player); err != nil {
            return err
        }
    case mafiapb.Action_PLAYER_NICKS:
        getNicksReq := action.GetPlayerNicks()
        player := s.playerById[getNicksReq.GetUserId()]
        if err := s.handlePlayerNicksRequest(player); err != nil {
            return err
        }
    case mafiapb.Action_VOTE:
        voteReq := action.GetVote()
        player := s.playerById[voteReq.GetUserId()]
        if err := s.handleVoteRequest(player, voteReq.GetMafiaUsername()); err != nil {
            return err
        }
    case mafiapb.Action_KILL_PLAYER_BY_MAFIA:
        killPlayerReq := action.GetKillPlayerByMafia()
        player := s.playerById[s.idByNickname[killPlayerReq.GetPlayerUsername()]]
        if err := s.handleMafiaKillPlayerRequest(killPlayerReq.GetUserId(), player); err != nil {
            return err
        }
    case mafiapb.Action_INVESTIGATE_MAFIA:
        investigateMafiaReq := action.GetInvestiageMafia()
        player := s.playerById[s.idByNickname[investigateMafiaReq.GetMafiaUsername()]]
        if err := s.handleInvestigateMafiaRequest(investigateMafiaReq.GetUserId(), player); err != nil {
            return err
        }
    case mafiapb.Action_INTERRUPT_GAME:
        interruptGameReq := action.GetInterruptGame()
        player := s.playerById[interruptGameReq.GetUserId()]
        if err := s.handleInterruptGameRequest(player); err != nil {
            return err
        }
    case mafiapb.Action_INVESTIGATION_RESULT:
        investigationResult := action.GetInvestigationResult()
        commissar := s.playerById[investigationResult.GetUserId()]
        if err := s.handleInvestigationResultEvent(commissar, investigationResult.GetPublishResult()); err != nil {
            return err
        }
    }

    return nil
}
