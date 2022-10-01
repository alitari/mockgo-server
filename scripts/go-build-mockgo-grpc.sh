#!/bin/sh

cd mockgo-grpc/cmd
CGO_ENABLED=0 GOOS=linux go build -v -o ../../bin/mockgo-grpc
echo "executable file:"
ls -l ../../bin/mockgo-grpc
cd -
