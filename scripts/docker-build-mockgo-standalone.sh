#!/bin/bash

set -e

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1
push="${2:-false}"

docker build -f build/docker/mockgo-standalone.Dockerfile . -t alitari/mockgo-standalone:$releaseTag --no-cache
trivy image alitari/mockgo-standalone:$releaseTag --format sarif --severity 'CRITICAL,HIGH' --output mockgo-standalone-trivy-results.sarif
if [ "$push" == "true" ]
then
    docker push alitari/mockgo-standalone:$releaseTag
fi