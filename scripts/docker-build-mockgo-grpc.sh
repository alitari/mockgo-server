#!/bin/sh

set -e

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1

docker build -f build/docker/mockgo-grpc.Dockerfile . -t alitari/mockgo-grpc:$releaseTag
docker push alitari/mockgo-grpc:$releaseTag
