#!/bin/sh

MINIKUBE_IP=$(minikube ip)
CONFIG_HOST=config.${MINIKUBE_IP}.nip.io
MOCK_HOST=mock.${MINIKUBE_IP}.nip.io
curl -v -f http://${CONFIG_HOST}/health