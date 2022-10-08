#!/bin/bash

set -e

MINIKUBE_IP=$(minikube ip)
STANDALONE_HOST=mockgo-standalone.${MINIKUBE_IP}.nip.io
GRPC_HOST=mockgo-grpc.${MINIKUBE_IP}.nip.io
curl -v -f http://${STANDALONE_HOST}/__/health
curl -v -f http://${GRPC_HOST}/__/health