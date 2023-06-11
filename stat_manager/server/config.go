package server

import "time"

type ServerConfig struct {
	Port string
	ReadTimeout time.Duration
	WriteTimeout time.Duration

	DatabaseFile string
}
