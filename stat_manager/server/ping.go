package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
)

const (
	pingRoute = "/ping"
)

func (sm *statManager) registerPingRoute() {
	sm.router.GET(pingRoute, func(c *gin.Context) {
		zlog.Info().Msg("accepted ping request")
		c.String(http.StatusOK, "This is stat manage!r\n")
	})
}
