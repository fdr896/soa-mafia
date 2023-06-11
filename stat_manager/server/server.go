package server

import (
	"common"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stat_manager/storage/database"
	"stat_manager/storage/filesystem"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
)

type statManager struct {
	// server
	instance *http.Server
	router   *gin.Engine

	// storage
	db       *database.Players
	as       *filesystem.AvatarsStorage
}

func NewStatManager(config *ServerConfig) (*statManager, error) {
	var sm statManager

	// Gin router
	sm.router = gin.New()

	sm.router.Use(gin.Logger())
	sm.router.Use(gin.Recovery())

	sm.registerPingRoute()
	sm.registerPlayersRoutes()

	// HTTP server
	sm.instance = &http.Server{
		Addr: ":" + config.Port,
		Handler: sm.router,
		ReadTimeout: config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	// database
	db, err := database.CreateOrReadPlayersDB(config.DatabaseFile)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to create db")
		return nil, err
	}
	sm.db = db

	as, err := filesystem.CreateAvatarsStorage()
	if err != nil {
		zlog.Error().Err(err).Msg("failed to create avatars storage")
		return nil, err
	}
	sm.as = as

	return &sm, nil
}

func (sm *statManager) Start() {
	common.InitServerLogger()

	go func() {
		fmt.Printf("StatManager started listening on %s.\nPress Ctrl+C to shutdown...\n", sm.instance.Addr)

		if err := sm.instance.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msg("failed when serving")
		}
	}()

	signalListener := make(chan os.Signal, 1)
	signal.Notify(signalListener, syscall.SIGINT, syscall.SIGTERM)
	<-signalListener
	fmt.Println("\nGracefully shutting down with timeout 5s")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := sm.instance.Shutdown(ctx); err != nil {
		zlog.Fatal().Err(err).Msg("failed to shutdown")
	}

	select {
	case <-ctx.Done():
		fmt.Println("Timeout ended! Force shuting down")
	default:
	}
}
