#!/bin/bash

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1
gitsha=$(git rev-parse --short HEAD)

docker login

set -e

PATH="$PATH:$(go env GOPATH)/bin" 

# create executabels
rm -f ./bin/*
for target in amd64 arm64
do
    ./scripts/go-build-mockgo.sh linux $target standalone ${releaseTag}-${gitsha}
done
./scripts/go-build-mockgo.sh windows amd64 standalone ${releaseTag}-${gitsha}

./scripts/go-build-grpc-matchstore.sh
./scripts/go-build-grpc-kvstore.sh

for target in amd64 arm64
do
./scripts/go-build-mockgo.sh linux $target grpc ${releaseTag}-${gitsha}
done
./scripts/go-build-mockgo.sh windows amd64 grpc ${releaseTag}-${gitsha}

# tgz
for file in ./bin/*
do
    sha256sum ${file} > ${file}.sha256
    tar -cvzf ${file}.tgz ${file}
    sha256sum --check ${file}.sha256
    rm ${file}
done

# login in github
gh auth login --with-token < .github/token
gh auth status
gh config set prompt disabled

# create release with tgz as assets
gh release create $releaseTag ./bin/*.*
gh auth logout -h github.com

# docker builds

./scripts/docker-build-mockgo.sh $releaseTag standalone true

./scripts/docker-build-mockgo.sh $releaseTag grpc true
