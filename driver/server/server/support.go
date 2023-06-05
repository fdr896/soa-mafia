package server

import "driver/server/game"

func (s *Server) findOrCreatedNotStartedSession() *game.GameSession {
	for session := range s.sessions {
		if !session.IsStarted() {
			return session
		}
	}

	session := game.NewGameSession(len(s.sessions))
	s.sessions[session] = struct{}{}

	return session
}
