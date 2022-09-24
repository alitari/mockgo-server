#!/bin/sh

count=$1
concurrent=$2


MINIKUBE_IP=$(minikube ip)
CONFIG_HOST=config.${MINIKUBE_IP}.nip.io
MOCK_HOST=mock.${MINIKUBE_IP}.nip.io

curl -u mockgo:asecretPassword -v -H "Accept: application/json" -X DELETE http://${CONFIG_HOST}/matches

hey -n $count -c $concurrent -m GET http://${MOCK_HOST}/hello9/some/stuff/sender/alex/additional/stuff/receiver/dani

response=$(curl -s -u mockgo:asecretPassword -H "Accept: application/json" http://${CONFIG_HOST}/matches)

expected="{\"hello9\":$count}"
if [ "$response" != "$expected" ]
then
    echo "failed expected '$expected' , but is '$response'"
    exit 1
else
    echo "success"
fi