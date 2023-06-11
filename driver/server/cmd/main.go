package main

import (
	"common"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	mafiapb "driver/server/proto"
	"driver/server/server"
	stat_manager "stat_manager/client"

	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	SESSION_PLAYERS_DEFAULT   = 4
	MAFIAS_DEFAULT = 4

	STAT_MANAGER_HOST_DEFAULT = "localhost"
	STAT_MANAGER_PORT_DEFAULT = "9077"
)

 func main() {
	rand.Seed(time.Now().UnixNano())

	port := os.Getenv("PORT")

	var sessionPlayers int
	if envSessionPlayers, set := os.LookupEnv("SESSION_PLAYERS"); set {
		sessionPlayersNum, err := strconv.Atoi(envSessionPlayers)
		if err != nil {
			log.Fatalln(err)
		}
		if sessionPlayersNum < 4 {
			log.Fatal("Minimum number of players is 4")
		}
		sessionPlayers = sessionPlayersNum
	} else {
		sessionPlayers = SESSION_PLAYERS_DEFAULT
	}

	var mafias int
	if envMafias, set := os.LookupEnv("MAFIAS"); set {
		mafiasNum, err := strconv.Atoi(envMafias)
		if err != nil {
			log.Fatalln(err)
		}
		if mafiasNum >= sessionPlayers - mafiasNum - 1 {
			log.Fatal("There are should be not more than (#sessionPlayers - #mafias - 1) mafias")
		}
		mafias = mafiasNum
	} else {
		mafias = MAFIAS_DEFAULT
	}

	common.InitServerLogger()

	statManagerClient := stat_manager.NewStatClient(
		common.GetEnvOrDefault("STAT_MANAGER_HOST", STAT_MANAGER_HOST_DEFAULT),
		common.GetEnvOrDefault("STAT_MANAGER_PORT", STAT_MANAGER_PORT_DEFAULT),
	)

	zlog.Info().
	     Str("port", port).
		 Int("session players", sessionPlayers).
		 Int("mafias", mafias).
		 Msg("running mafia driver server")
	listener, err := net.Listen("tcp", ":" + port)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to start listening")
	}


	grpcServer := grpc.NewServer()
	mafiapb.RegisterMafiaDriverServer(
		grpcServer, server.NewServer(sessionPlayers, mafias, statManagerClient))

    zlog.Info().Str("port", port).Msg("server listening")
	if err := grpcServer.Serve(listener); err != nil {
		zlog.Fatal().Err(err).Msg("failed to start server")
	}
 }

