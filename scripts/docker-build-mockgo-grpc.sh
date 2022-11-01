#!/bin/bash

set -e

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1
push="${2:-false}"

docker build -f build/docker/mockgo-grpc.Dockerfile . -t alitari/mockgo-grpc:$releaseTag --no-cache
trivy image alitari/mockgo-grpc:$releaseTag --format sarif --severity 'CRITICAL,HIGH' --output mockgo-grpc-trivy-results.sarif
cat mockgo-grpc-trivy-results.sarif
if [ "$push" == "true" ]
then
    docker push alitari/mockgo-grpc:$releaseTag
fi
