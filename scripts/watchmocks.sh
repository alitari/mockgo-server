#!/bin/bash

function stop_mockgo {
    echo "Stopping mockgo..."
    kill $pid
    exit
}

set -e
if [[ -z $1 ]]; then
    echo "no mock directory supplied"
    exit 1
fi
mockdir=$1
if [[ -z $2 ]]; then
    mockgo_executable="./mockgo-standalone/cmd/bin/mockgo-standalone-linux-amd64"
else
    mockgo_executable=$2
fi

trap stop_mockgo SIGINT

echo "Start mockgo..."
MOCK_DIR=$mockdir $mockgo_executable &
pid=$!

sleep 1
echo "Press Ctrl-C to stop the command."

while true; do
    inotifywait -e modify,create,delete -r $mockdir && \
    curl -X POST -u mockgo:password localhost:8081/__/reload
done


