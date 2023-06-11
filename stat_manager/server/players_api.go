package server

import (
	"fmt"
	"net/http"
	"net/mail"
	"stat_manager/storage/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

const (
	playerRoute = "/player"
)

func (sm *statManager) registerPlayersRoutes() {
	router := sm.router

	//////////// GET ////////////
	router.GET(playerRoute+"/:username", sm.handleGetOnePlayerInfoRequest)
	router.GET(playerRoute, sm.handleGetManyPlayersInfoRequest)
	router.GET(playerRoute+"/:username/avatar", sm.handleGetAvatarRequest)
	router.GET(playerRoute+"/:username/pdf", sm.handleGetPlayerPdfRequest)

	//////////// POST ////////////
	router.POST(playerRoute, sm.handleCreatePlayerRequest)
	router.POST(playerRoute+"/:username/pdf", sm.handleSubmitGenPdfRequest)

	//////////// PUT ////////////
	router.PUT(playerRoute+"/:username", sm.handleUpdatePlayerRequest)

	//////////// DELETE ////////////
	router.DELETE(playerRoute+"/:username", sm.handleDeletePlayerRequest)
}

func (sm *statManager) handleGetOnePlayerInfoRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().Str("username", username.Username).Msg("get player info request")

	p, err := sm.db.GetPlayerByUsername(c, username.Username)
	switch err {
	case nil:
		zlog.Debug().Interface("player", p).Msg("queried")
		c.JSON(http.StatusOK, fromDbPlayer(p))
	case database.ErrNotFound:
		zlog.Error().Str("username", username.Username).Msg("not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "no player with given username"})
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (sm *statManager) handleGetManyPlayersInfoRequest(c *gin.Context) {
	usernames := c.Query("usernames")
	if usernames == "" {
		zlog.Error().Msg("usernames param not specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "specify 'usernames' query param"})
		return
	}

	zlog.Info().Str("usernames", usernames).Msg("get many players info request")

	usernamesArray := strings.Split(usernames, ",")

	ps, err := sm.db.GetPlayersByUsernames(c, usernamesArray)
	switch err {
	case nil:
		zlog.Debug().Interface("players", ps).Msg("queried")
		c.JSON(http.StatusOK, fromDbPlayers(ps))
	case database.ErrNotFound:
		zlog.Error().Str("usernames", usernames).Msg("not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "some usernames are invalid"})
	default:
		zlog.Error().Err(err).Msg("failed to query players")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (sm *statManager) handleGetAvatarRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().Str("username", username.Username).Msg("get avatar request")

	p, err := sm.db.GetPlayerByUsername(c, username.Username)
	switch err {
	case nil:
		zlog.Debug().Interface("player", p).Msg("queried")
		c.File(sm.as.GetAvatarPath(p.AvatarFilename))
	case database.ErrNotFound:
		zlog.Error().Str("username", username.Username).Msg("not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "no player with given username"})
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (sm *statManager) handleCreatePlayerRequest(c *gin.Context) {
	var p Player
	if err := c.Bind(&p); err != nil {
		zlog.Error().Err(err).Msg("failed to bind player")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().
		Str("username", p.Username).
		Str("email", p.Email).
		Str("gender", p.Gender).
		Bool("has avatar file", p.Avatar != nil).
		Msg("create player request")
	
	_, err := sm.db.GetPlayerByUsername(c, p.Username)
	switch err {
	case database.ErrNotFound:
		zlog.Info().Msg("player will be created")
	case nil:
		zlog.Error().Str("username", p.Username).Msg("player already exist")
		c.JSON(http.StatusConflict, gin.H{"error": "player with same username already exists"})
		return
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dbPlayer := func() *database.Player {
		if p.Avatar == nil {
			zlog.Info().Msg("creating player with default avatar")
			return toDbPlayer(&p)
		} else {
			avatarFile, err := c.FormFile("avatar")
			if err != nil {
				zlog.Error().Err(err).Msg("failed to retrieve file from request")
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "failed to retrieve file from request")})
				return nil
			}
			avatar, err := avatarFile.Open()
			if err != nil {
				zlog.Error().Err(err).Msg("failed to open avatar file")
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "failed to open avatar file")})
				return nil
			}
			avatarPath, err := sm.as.WriteUserAvatar(p.Username, avatarFile.Header.Get("Content-Type"), avatar);
			if err != nil {
				zlog.Error().Err(err).Msg("failed to copy avatar file")
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "failed to read avatar file")})
				return nil
			}
			zlog.Info().Msg("creating player with avatar")
			return toDbPlayerWithAvatar(&p, avatarPath)
		}
	}()
	if dbPlayer == nil {
		return
	}

	if err := sm.db.CreatePlayer(c, dbPlayer); err != nil {
		zlog.Error().Err(err).Msg("failed to create player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, "player created\n")
}

func (sm *statManager) handleUpdatePlayerRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p PlayerForUpdate
	if err := c.Bind(&p); err != nil {
		zlog.Error().Err(err).Msg("failed to bind player for update")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if p.Email != nil {
		if _, err := mail.ParseAddress(*p.Email); err != nil {
			zlog.Error().Err(err).Msg("failed to parse email")
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "invalid email").Error()})
			return
		}
	}
	if p.Gender != nil &&
	   *p.Gender != database.GenderToString(database.MALE) &&
	   *p.Gender != database.GenderToString(database.FEMALE) {
		zlog.Error().Str("gender", *p.Gender).Msg("invalid gender")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid gender: gender must be one of 'male' or 'female'"})
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
		Interface("player", p).
		Bool("has avatar file", p.Avatar != nil).
		Msg("update player request")

	dbPlayer := func() *database.Player {
		if p.Avatar == nil {
			zlog.Info().Msg("updating only personal info")
			return toDbPlayerForUpdate(username.Username, &p)
		} else {
			avatarFile, err := c.FormFile("avatar")
			if err != nil {
				zlog.Error().Err(err).Msg("failed to retrieve file from request")
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "failed to retrieve file from request")})
				return nil
			}
			avatar, err := avatarFile.Open()
			if err != nil {
				zlog.Error().Err(err).Msg("failed to open avatar file")
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "failed to open avatar file")})
				return nil
			}
			avatarPath, err := sm.as.WriteUserAvatar(username.Username, avatarFile.Header.Get("Content-Type"), avatar);
			if err != nil {
				zlog.Error().Err(err).Msg("failed to copy avatar file")
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "failed to read avatar file")})
				return nil
			}
			zlog.Info().Msg("updating avatar")
			return toDbPlayerForUpdateWithAvatar(username.Username, &p, avatarPath)
		}
	}()
	if dbPlayer == nil {
		return
	}
	
	if err := sm.db.UpdatePlayer(c, dbPlayer); err != nil {
		zlog.Error().Err(err).Msg("failed to update player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (sm *statManager) handleDeletePlayerRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().Str("username", username.Username).Msg("delete player request")

	err := sm.db.DeletePlayerByUsername(c, username.Username)
	switch err {
	case nil:
		zlog.Info().Str("username", username.Username).Msg("deleted")
		c.JSON(http.StatusNoContent, gin.H{})
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (sm *statManager) handleGetPlayerPdfRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := sm.db.GetPlayerByUsername(c, username.Username)
	switch err {
	case nil:
		zlog.Info().Msg("pdf will be checked")
	case database.ErrNotFound:
		zlog.Error().Str("username", username.Username).Msg("no such player")
		c.JSON(http.StatusNotFound, gin.H{"error": "player with given username does not exists"})
		return
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().Str("username", username.Username).Msg("get pdf request")

	if sm.ps.Exists(username.Username) {
		c.File(sm.ps.UserPdfPath(username.Username))
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "pdf is not generated yet"})
	}
}

func (sm *statManager) handleSubmitGenPdfRequest(c *gin.Context) {
	var username PlayerUsername
	if err := c.BindUri(&username); err != nil {
		zlog.Error().Err(err).Msg("failed to bind username")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zlog.Info().Str("username", username.Username).Msg("get pdf gen request")

	p, err := sm.db.GetPlayerByUsername(c, username.Username)
	switch err {
	case nil:
		zlog.Info().Msg("task will be submited")
	case database.ErrNotFound:
		zlog.Error().Str("username", username.Username).Msg("no such player")
		c.JSON(http.StatusNotFound, gin.H{"error": "player with given username does not exists"})
		return
	default:
		zlog.Error().Err(err).Msg("failed to query player")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := sm.tm.SubmitPdfGenTask(p); err != nil {
		zlog.Error().Err(err).Msg("failed to submit gen pdf task")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, sm.genPdfEndpoint(p.Username))
}

func (sm *statManager) genPdfEndpoint(username string) string {
	return fmt.Sprintf("%s/player/%s/pdf\n", sm.getEndpoint(), username)
}
