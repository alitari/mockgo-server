#!/bin/bash

set -e

cd mockgo-grpc/cmd

for target in amd64 arm64
do
    CGO_ENABLED=0 GOARCH=$target GOOS=linux go build -v -o ../../bin/mockgo-grpc-linux-${target}
done

CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -v -o ../../bin/mockgo-grpc-windows-amd64

echo "executable files:"
ls -l ../../bin/mockgo-grpc*
cd -
