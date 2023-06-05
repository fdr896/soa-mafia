package main

import (
	"context"
	"driver/client/client"
	"fmt"
	"log"

	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    if len(os.Args) != 4 {
        log.Fatalf("Usage: go run client/cmd/main.go [manual|auto] <username> <port>\n")
    }
    mode := os.Args[1]
    if mode != "manual" && mode != "auto" {
        fmt.Println("Wrong mode")
        log.Fatalf("Usage: go run client/cmd/main.go [manual|auto] <username> <port>\n")
    }
    username := os.Args[2]
    port := os.Args[3]

    fmt.Printf("connecting to grpc server by port [%s]...\n", port)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
        log.Fatalf("failed to connect to grpc server: %s\n", err.Error())
	}
	defer conn.Close()
    fmt.Printf("connected to grpc server by port [%s]\n", port)

    client, err  := client.NewClient(mode, username, conn)
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
