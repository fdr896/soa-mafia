#!/bin/bash

docker build -t fdr400/soa_mafia_driver_client -f client/cmd/Dockerfile .
docker build -t fdr400/soa_mafia_driver_server -f server/cmd/Dockerfile .
