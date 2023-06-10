#!/bin/bash

docker build -t fdr400/soa_mafia_driver_client:testing -f driver/client/cmd/Dockerfile .
docker build -t fdr400/soa_mafia_driver_server:testing -f driver/server/cmd/Dockerfile .

docker image tag soa_mafia_driver_client:testing fdr400/soa_mafia_driver_client:testing
docker image tag soa_mafia_driver_server:testing fdr400/soa_mafia_driver_server:server

docker image push fdr400/soa_mafia_driver_client:testing
docker image push fdr400/soa_mafia_driver_server:testing
