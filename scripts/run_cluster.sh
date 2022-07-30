#!/bin/bash

cd cmd
for PORT in 8080 8082 8084
do
    MOCK_PORT=$PORT CONFIG_PORT=$((PORT+1)) MOCK_DIR=. CLUSTER_URLS="http://localhost:8081,http://localhost:8083,http://localhost:8085" go run . &
done
