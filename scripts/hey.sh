#!/bin/sh

count=$1
concurrent=$2
quote=$3
host=$4


MINIKUBE_IP=$(minikube ip)
MOCKGO_HOST=${host}.${MINIKUBE_IP}.nip.io

echo "Running load tests for target host '$MOCKGO_HOST' ..."

curl -u mockgo:asecretPassword -v -H "Accept: application/json" -X DELETE http://${MOCKGO_HOST}/__/matches/hello9

hey -n $count -c $concurrent -q $quote -m GET http://${MOCKGO_HOST}/hello9/some/stuff/sender/alex/additional/stuff/receiver/dani

response=$(curl -s -u mockgo:asecretPassword -H "Accept: application/json" http://${MOCKGO_HOST}/__/matchesCount/hello9)

expected="$count"
if [ "$response" != "$expected" ]
then
    echo "failed expected '$expected' , but is '$response'"
    exit 1
else
    echo "success"
fi