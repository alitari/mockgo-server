#!/bin/bash
set -e

if [ $# -lt 4 ]
then
    echo "Need 4 args: os, arch, variant, releaseTag"
    exit 1
fi

os=$1
arch=$2
variant=$3
releaseTag=$4

echo "building mockgo-$variant $releaseTag ..."

cd "mockgo-${variant}/cmd"
sed -i "s/const versionTag = .*/const versionTag = \"${releaseTag}\"/g" main.go

CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -v -o ../../bin/mockgo-${variant}-${os}-${arch}

echo "executable file:"
ls -l ../../bin/mockgo-${variant}-${os}-${arch}
cd -
