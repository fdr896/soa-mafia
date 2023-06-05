package main

import (
	"driver/support"
	"net"

	mafiapb "driver/server/proto"
	"driver/server/server"

	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

 func main() {
	support.InitServerLogger()

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to start listening")
	}

	grpcServer := grpc.NewServer()
	mafiapb.RegisterMafiaDriverServer(grpcServer, server.NewServer())

    zlog.Info().Str("port", "9000").Msg("server listening")
	if err := grpcServer.Serve(listener); err != nil {
		zlog.Fatal().Err(err).Msg("failed to start server")
	}
 }
