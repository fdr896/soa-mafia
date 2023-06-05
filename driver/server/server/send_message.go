package server

import (
	"driver/server/game"
	mafiapb "driver/server/proto"

	zlog "github.com/rs/zerolog/log"
)

func (s *Server) sendMessageToPlayer(p *player, msg *mafiapb.ActionResponse) error {
	playerStream := s.playerById[p.id].stream
    zlog.Info().Str("nick", p.id).Str("msg", msg.String()).Msg("sending direct msg")
	return playerStream.Send(msg)
}

func (s *Server) sendMessageToGameSession(gameSession *game.GameSession, msg *mafiapb.ActionResponse) error {
	sessionPlayers := s.sessionPlayers[gameSession]
	for _, player := range sessionPlayers {
        zlog.Info().Str("id", player.id).Str("msg", msg.String()).Msg("sending broadcast msg")
		if err := player.stream.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

type GenActionResponse func(*player, *game.GameSession) *mafiapb.ActionResponse

func (s *Server) sendMessageToGameSessionWithCb(gameSession *game.GameSession, cb GenActionResponse) error {
	sessionPlayers := s.sessionPlayers[gameSession]
	for _, player := range sessionPlayers {
        msg := cb(player, gameSession)
        zlog.Info().Str("id", player.id).Str("msg", msg.String()).Msg("sending generated broadcast msg")
		if err := player.stream.Send(msg); err != nil {
			return err
		}
	}
	return nil
}