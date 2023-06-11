package client

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"stat_manager/server"

	zlog "github.com/rs/zerolog/log"
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

	resp, err := http.Post(sc.serverEndpoint, "application/json", bytes.NewBuffer(statBytes))
	if err != nil {
		return err
	}

	zlog.Info().Str("status", resp.Status).Msg("update stat")

	return nil
}
