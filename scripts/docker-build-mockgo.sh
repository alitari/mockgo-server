#!/bin/bash

set -e

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1
variant="${2:-standalone}"
push="${3:-false}"
gitsha=$(git rev-parse --short HEAD)

docker build --build-arg RELEASE=${releaseTag}-${gitsha} -f build/docker/mockgo-${variant}.Dockerfile . -t alitari/mockgo-${variant}:$releaseTag --no-cache 
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/reports:/reports aquasec/trivy:0.34.0 image alitari/mockgo-${variant}:$releaseTag --format sarif --severity 'CRITICAL,HIGH' --output /reports/mockgo-${variant}-trivy-results.sarif
if [ "$push" == "true" ]
then
    docker push alitari/mockgo-${variant}:$releaseTag
fi