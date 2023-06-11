package main

import (
	"common"
	"log"
	stat_manager "stat_manager/server"
	"strconv"
	"time"
)

const (
	STAT_MANAGER_PORT_DEFAULT = "9077"
	STAT_MANAGER_READ_TIMEOUT_MS_DEFAULT = "5000"
	STAT_MANAGER_WRITE_TIMEOUT_MS_DEFAULT = "5000"
	STAT_MANAGER_DATABASE_FILE_DEFAULT = "players.db"

	RABBITMQ_USER_DEFAULT     = "guest"
	RABBITMQ_PASSWORD_DEFAULT = "guest"
	RABBITMQ_HOSTNAME_DEFAULT = "localhost"
	RABBITMQ_PORT_DEFAULT     = "5672"
)

func main() {
	readTimeout, err := strconv.Atoi(
		common.GetEnvOrDefault("STAT_MANAGER_READ_TIMEOUT_MS", STAT_MANAGER_READ_TIMEOUT_MS_DEFAULT))
	if err != nil {
		log.Fatalln(err)
	}
	writeTimeout, err := strconv.Atoi(
		common.GetEnvOrDefault("STAT_MANAGER_WRITE_TIMEOUT_MS", STAT_MANAGER_WRITE_TIMEOUT_MS_DEFAULT))
	if err != nil {
		log.Fatalln(err)
	}

	config := &stat_manager.ServerConfig{
		Port: common.GetEnvOrDefault("STAT_MANAGER_PORT", STAT_MANAGER_PORT_DEFAULT),
		ReadTimeout: time.Duration(readTimeout * int(time.Millisecond)),
		WriteTimeout: time.Duration(writeTimeout * int(time.Millisecond)),
		DatabaseFile: common.GetEnvOrDefault("STAT_MANAGER_DATABASE_FILE", STAT_MANAGER_DATABASE_FILE_DEFAULT),
	}

	rabbitmqConnParams := common.NewRabbitmqConnectionParams(
		common.GetEnvOrDefault("RABBITMQ_USER", RABBITMQ_USER_DEFAULT),
		common.GetEnvOrDefault("RABBITMQ_PASSWORD", RABBITMQ_PASSWORD_DEFAULT),
		common.GetEnvOrDefault("RABBITMQ_HOSTNAME", RABBITMQ_HOSTNAME_DEFAULT),
		common.GetEnvOrDefault("RABBITMQ_PORT", RABBITMQ_PORT_DEFAULT),
	)

	statManager, err := stat_manager.NewStatManager(config, rabbitmqConnParams)
	if err != nil {
		log.Fatalf("failed to create stat manager: %v", err)
	}

	statManager.Start()
}
