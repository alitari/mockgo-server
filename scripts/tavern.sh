#!/bin/sh

TAVERNVER=1.23.3
docker build --build-arg TAVERNVER=$TAVERNVER --file test/tavern/tavern.Dockerfile --tag tavern:$TAVERNVER test/tavern

MINIKUBE_IP=$(minikube ip)
CONFIG_HOST=config.${MINIKUBE_IP}.nip.io
MOCK_HOST=mock.${MINIKUBE_IP}.nip.io

docker run --network host -e CONFIG_HOST=$CONFIG_HOST -e MOCK_HOST=$MOCK_HOST -v ${PWD}/test/tavern:/tavern -v ${PWD}/reports:/reports tavern:$TAVERNVER py.test -vv --html=/reports/tavern.html --self-contained-html /tavern/test.tavern.yaml