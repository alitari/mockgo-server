#!/bin/bash
set -e

if [ $# -lt 2 ]
then
    echo "Need 2 args, os and arch"
    exit 1
fi

os=$1
arch=$2

cd mockgo-grpc/cmd

CGO_ENABLED=0 GOOS=$os GOARCH=$arch  go build -v -o ../../bin/mockgo-grpc-${os}-${arch}

echo "executable files:"
ls -l ../../bin/mockgo-grpc*
cd -
