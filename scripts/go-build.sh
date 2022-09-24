#!/bin/sh

cd mockgo-server/cmd
CGO_ENABLED=0 GOOS=linux go build -v -o ../../bin/mockgo-server
echo "executable file:"
ls -l ../bin/mockgo-server
cd -
