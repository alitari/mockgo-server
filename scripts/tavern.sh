#!/bin/sh

execute_tavern () {
    echo "Executing tests for target host: '$1'"
    docker run --network host -e MOCKGO_HOST=$1 -v ${PWD}/test/tavern:/tavern -v ${PWD}/reports:/reports tavern:$TAVERNVER py.test -vv --html=/reports/tavern.html --self-contained-html /tavern/test.tavern.yaml
}


TAVERNVER=1.23.3
docker build --build-arg TAVERNVER=$TAVERNVER --file test/tavern/tavern.Dockerfile --tag tavern:$TAVERNVER test/tavern
MINIKUBE_IP=$(minikube ip)

execute_tavern mockgo-standalone.${MINIKUBE_IP}.nip.io

execute_tavern mockgo-grpc.${MINIKUBE_IP}.nip.io

