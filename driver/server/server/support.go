package server

import (
	"driver/server/game"

	"github.com/beevik/guid"
)

func (s *Server) findOrCreatedNotStartedSession() *game.GameSession {
	for session := range s.sessions {
		if !session.IsStarted() {
			return session
		}
	}

	session := game.NewGameSession("sess_" + guid.NewString(), s.gamePlayers, s.mafias)
	s.sessions[session] = struct{}{}

	return session
}
