package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"stat_manager/server"

	zlog "github.com/rs/zerolog/log"
)

const (
	updateStatRoute = "internal/player"
)

type StatClient struct {
	serverEndpoint string
	client *http.Client
}

func NewStatClient(host, port string) *StatClient {
	return &StatClient{
		serverEndpoint: "http://" + net.JoinHostPort(host, port),
		client: http.DefaultClient,
	}
}

func (sc *StatClient) UpdatePlayerStat(stat *server.PlayerStat) error {
	statBytes, err := json.Marshal(*stat)
	if err != nil {
		return err
	}

	route := sc.getUserUpdateStatRoute(stat)
	zlog.Debug().Str("route", route).Msg("sending request")
	resp, err := http.Post(route, "application/json", bytes.NewBuffer(statBytes))
	if err != nil {
		return err
	}

	zlog.Info().Str("status", resp.Status).Msg("update stat")

	return nil
}

func (sc *StatClient) getUserUpdateStatRoute(stat *server.PlayerStat) string {
	return fmt.Sprintf("%s/%s/%s", sc.serverEndpoint, updateStatRoute, stat.Username)
}
