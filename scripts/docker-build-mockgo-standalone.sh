#!/bin/sh

set -e

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1

docker build -f build/docker/mockgo-standalone.Dockerfile . -t alitari/mockgo-standalone:$releaseTag
trivy image alitari/mockgo-standalone:$releaseTag --severity 'CRITICAL,HIGH' --exit-code 1
docker push alitari/mockgo-standalone:$releaseTag