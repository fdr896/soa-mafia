package main

import (
	"chat"
	"common"
	"context"
	"driver/client/client"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CLIENT_MODE_DEFAULT 	  = "manual"
	SERVER_HOST_DEFAULT       = "localhost"
	SERVER_PORT_DEFAULT       = "9000"

	RABBITMQ_USER_DEFAULT     = "guest"
	RABBITMQ_PASSWORD_DEFAULT = "guest"
	RABBITMQ_HOSTNAME_DEFAULT = "localhost"
	RABBITMQ_PORT_DEFAULT     = "5672"
)

func main() {
    rand.Seed(time.Now().UnixNano())

	mode := common.GetEnvOrDefault("CLIENT_MODE", CLIENT_MODE_DEFAULT)
    if mode != "manual" && mode != "auto" {
        fmt.Println("Wrong mode")
        log.Fatalln("Set up CLIENT_MODE to 'manual' or 'auto'")
    }

    username := os.Getenv("USERNAME")
    if len(username) == 0 && mode == "auto" {
        log.Fatalln("Set up not empty USERNAME for bot")
    }
    serverPort := common.GetEnvOrDefault("SERVER_PORT", SERVER_PORT_DEFAULT)
    serverHost := common.GetEnvOrDefault("SERVER_HOST", SERVER_HOST_DEFAULT)
    if len(serverHost) == 0 {
        serverHost = "localhost"
    }

    serverEndpoint := net.JoinHostPort(serverHost, serverPort)
    fmt.Printf("connecting to grpc server by endpoint [%s]...\n", serverEndpoint)
	conn, err := grpc.Dial(
		serverEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
        log.Fatalf("failed to connect to grpc server: %s\n", err.Error())
	}
	defer conn.Close()
    fmt.Printf("connected to grpc server on endpoint [%s]\n", serverEndpoint)

	rabbitmqConnParams := chat.NewRabbitmqConnectionParams(
		common.GetEnvOrDefault("RABBITMQ_USER", RABBITMQ_USER_DEFAULT),
		common.GetEnvOrDefault("RABBITMQ_PASSWORD", RABBITMQ_PASSWORD_DEFAULT),
		common.GetEnvOrDefault("RABBITMQ_HOSTNAME", RABBITMQ_HOSTNAME_DEFAULT),
		common.GetEnvOrDefault("RABBITMQ_PORT", RABBITMQ_PORT_DEFAULT),
	)

    client, err  := func() (client.IClient, error) {
		switch mode {
		case "manual":
			return client.NewManualClient(mode, conn, rabbitmqConnParams)
		case "auto":
			return client.NewAutoClient(mode, username, conn, rabbitmqConnParams)
		default:
			panic("bit flipped!")
		}
	}()
    if err != nil {
        log.Fatalf("failed to create client: %s\n", err.Error())
    }
    if err := client.OpenStream(context.Background()); err != nil {
        log.Fatalf("failed to open stream with server: %s\n", err.Error())
    }

    if err := client.StartPlaying(); err != nil {
        log.Fatalf("error occured when playing: %s\n", err.Error())
    }
}
