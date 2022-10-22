#!/bin/bash
set -e

if [ $# -lt 2 ]
then
    echo "Need 2 args, os and arch"
    exit 1
fi

os=$1
arch=$2

cd mockgo-standalone/cmd

CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -v -o ../../bin/mockgo-standalone-${os}-${arch}

echo "executable file:"
ls -l ../../bin/mockgo-standalone-${os}-${arch}
cd -
