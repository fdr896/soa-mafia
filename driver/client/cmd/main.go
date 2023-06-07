package main

import (
	"chat"
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
	RABBITMQ_USER_DEFAULT     = "guest"
	RABBITMQ_PASSWORD_DEFAULT = "guest"
	RABBITMQ_HOSTNAME_DEFAULT = "localhost"
	RABBITMQ_PORT_DEFAULT     = "5672"
)

func main() {
    rand.Seed(time.Now().UnixNano())

    mode := os.Getenv("CLIENT_MODE")
    if mode != "manual" && mode != "auto" {
        fmt.Println("Wrong mode")
        log.Fatalln("Set up CLIENT_MODE to 'manual' or 'auto'")
    }
    username := os.Getenv("USERNAME")
    if len(username) == 0 {
        log.Fatalln("Set up not empty USERNAME")
    }
    serverPort := os.Getenv("SERVER_PORT")
    serverHost := os.Getenv("SERVER_HOST")
    if len(serverHost) == 0 {
        serverHost = "localhost"
    }

    serverEndpoint := net.JoinHostPort(serverHost, serverPort)
    fmt.Printf("connecting to grpc server by endpoint [%s]...\n", serverEndpoint)
	conn, err := grpc.Dial(serverEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
        log.Fatalf("failed to connect to grpc server: %s\n", err.Error())
	}
	defer conn.Close()
    fmt.Printf("connected to grpc server on endpoint [%s]\n", serverEndpoint)

	rabbitmqConnParams := chat.NewRabbitmqConnectionParams(
		getEnvOrDefault("RABBITMQ_USER", RABBITMQ_USER_DEFAULT),
		getEnvOrDefault("RABBITMQ_PASSWORD", RABBITMQ_PASSWORD_DEFAULT),
		getEnvOrDefault("RABBITMQ_HOSTNAME", RABBITMQ_HOSTNAME_DEFAULT),
		getEnvOrDefault("RABBITMQ_PORT", RABBITMQ_PORT_DEFAULT),
	)

    client, err  := client.NewClient(mode, username, conn, rabbitmqConnParams)
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

 func getEnvOrDefault(envName, defaultValue string) string {
	if envValue, set := os.LookupEnv(envName); set {
		return envValue
	} else {
		return defaultValue
	}
 }
