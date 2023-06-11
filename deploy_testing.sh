#!/bin/bash

docker build -t fdr400/soa_mafia_driver_client:testing -f driver/client/cmd/Dockerfile .
docker build -t fdr400/soa_mafia_driver_server:testing -f driver/server/cmd/Dockerfile .
docker build -t fdr400/soa_mafia_stat_manager:testing  -f stat_manager/cmd/Dockerfile .

docker image push fdr400/soa_mafia_driver_client:testing
docker image push fdr400/soa_mafia_driver_server:testing
docker image push fdr400/soa_mafia_stat_manager:testing
