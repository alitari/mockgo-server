#!/bin/sh

cd mockgo-standalone/cmd
CGO_ENABLED=0 GOOS=linux go build -v -o ../../bin/mockgo-standalone
echo "executable file:"
ls -l ../../bin/mockgo-standalone
cd -
