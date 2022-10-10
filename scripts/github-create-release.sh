#!/bin/bash

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1

set -e

PATH="$PATH:$(go env GOPATH)/bin" 

# create executabels
rm -f ./bin/*
./scripts/go-build-mockgo-standalone.sh
./scripts/go-build-grpc-matchstore.sh
./scripts/go-build-grpc-kvstore.sh
./scripts/go-build-mockgo-grpc.sh

# tgz
for file in ./bin/*
do
    tar -cvzf ${file}.tgz ${file}
done

# login in github
gh auth login --with-token < .github/token
gh auth status
gh config set prompt disabled

# create release with tgz as assets
gh release create $releaseTag ./bin/*.tgz

gh auth logout -h github.com