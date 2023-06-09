#!/bin/bash

docker build -t fdr400/soa_mafia_driver_client -f driver/client/cmd/Dockerfile .
docker build -t fdr400/soa_mafia_driver_server -f driver/server/cmd/Dockerfile .
docker build -t fdr400/soa_mafia_stat_manager  -f stat_manager/cmd/Dockerfile .

docker image push fdr400/soa_mafia_driver_client
docker image push fdr400/soa_mafia_driver_server
docker image push fdr400/soa_mafia_stat_manager
