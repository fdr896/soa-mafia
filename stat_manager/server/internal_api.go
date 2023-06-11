package server

import (
	"net/http"
	"stat_manager/storage/database"

	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
)

const (
	internalRoute = "/internal"
)

func (sm *statManager) registerInternalRoutes() {
	sm.router.POST(internalRoute+"/player/:username", sm.handleUpdatePlayerStatRequest)
}

func (sm *statManager) handleUpdatePlayerStatRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var s PlayerStat
	if err := c.Bind(&s); err != nil {
		zlog.Error().Err(err).Msg("failed to bind player stat")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := sm.db.GetPlayerByUsername(c, username.Username)
	switch err {
	case nil:
		zlog.Info().Msg("player will be updated")
	case database.ErrNotFound:
		zlog.Error().Str("username", username.Username).Msg("no such player")
		c.JSON(http.StatusNotFound, gin.H{"error": "player with given username does not exists"})
		return
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().
		Interface("player", s).
		Msg("update player stat request")
	
	dbPlayer := toDbPlayerStat(&s)

	if err := sm.db.UpdatePlayer(c, dbPlayer); err != nil {
		zlog.Error().Err(err).Msg("failed to update player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
