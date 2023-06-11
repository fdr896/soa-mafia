package server

import (
	"common"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stat_manager/pdf"
	"stat_manager/storage/database"
	"stat_manager/storage/filesystem"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
)

type statManager struct {
	// config
	config *ServerConfig

	// server
	instance *http.Server
	router   *gin.Engine

	// storage
	db       *database.Players
	as       *filesystem.AvatarsStorage
	ps       *filesystem.PdfStorage

	// pdf gen
	tm       *pdf.TaskManager
	pr       *pdf.PdfRender
}

func NewStatManager(config *ServerConfig, connParams *common.RabbitmqConnectionParams) (*statManager, error) {
	common.InitServerLogger()

	var sm statManager
	sm.config = config

	// Gin router
	sm.router = gin.New()

	sm.router.Use(gin.Logger())
	sm.router.Use(gin.Recovery())

	sm.registerPingRoute()
	sm.registerPlayersRoutes()
	sm.registerInternalRoutes()

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

	// file storages
	as, err := filesystem.CreateAvatarsStorage()
	if err != nil {
		zlog.Error().Err(err).Msg("failed to create avatars storage")
		return nil, err
	}
	sm.as = as

	ps, err := filesystem.CreatePdfStorage()
	if err != nil {
		zlog.Error().Err(err).Msg("failed to create pdf storage")
		return nil, err
	}
	sm.ps = ps

	// tasks manager
	tm := pdf.NewTaskManager(connParams)
	if err := tm.Start(); err != nil {
		zlog.Error().Err(err).Msg("failed to start tasks manager")
		panic(err)
	}
	sm.tm = tm

	// rendering
	sm.pr = pdf.NewRender(sm.tm, sm.ps, sm.as)
	go sm.pr.StartRendering()

	return &sm, nil
}

func (sm *statManager) Start() {
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

func (sm *statManager) getEndpoint() string {
	return fmt.Sprintf("http://localhost:%s", sm.config.Port)
}
