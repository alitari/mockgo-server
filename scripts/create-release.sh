#!/bin/bash

set -e

export MOCKGO_RELEASE=$1
# check if release tag is supplied
if [[ -z $MOCKGO_RELEASE ]]; then
    echo "no tag supplied, just testing a release run"
else
    if [[ ! $MOCKGO_RELEASE =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "release tag must be in semver format, e.g. v1.0.0"
        exit 1
    fi
    # git checks
    branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$branch" != "master" ]
    then
        echo "you must be in 'master' branch, but you are in '$branch'!"
        exit 1
    fi
#    gitstatus=$(git status --short)
#    if [[ "$gitstatus" != "" ]]; then
#        echo "the workspace is dirty: $gitstatus !"
#        exit 1
#    fi

    # create a release branch
    git checkout -b "release-$MOCKGO_RELEASE"
fi

echo "start release test 󰕹 ..."
make env

make clean MOCKGO_MODULE=mockgo
make tidy MOCKGO_MODULE=mockgo
make build MOCKGO_MODULE=mockgo
make cover MOCKGO_MODULE=mockgo

make dep-dev MOCKGO_MODULE=mockgo-standalone
make tidy MOCKGO_MODULE=mockgo-standalone
make clean MOCKGO_MODULE=mockgo-standalone
make cover MOCKGO_MODULE=mockgo-standalone
make acctest MOCKGO_MODULE=mockgo-standalone
make helm-delete MOCKGO_MODULE=mockgo-standalone

make dep-dev MOCKGO_MODULE=grpc-kvstore
make tidy MOCKGO_MODULE=grpc-kvstore
make clean MOCKGO_MODULE=grpc-kvstore
make build MOCKGO_MODULE=grpc-kvstore
make cover MOCKGO_MODULE=grpc-kvstore

make dep-dev MOCKGO_MODULE=grpc-matchstore
make tidy MOCKGO_MODULE=grpc-matchstore
make clean MOCKGO_MODULE=grpc-matchstore
make build MOCKGO_MODULE=grpc-matchstore
make cover MOCKGO_MODULE=grpc-matchstore

make dep-dev MOCKGO_MODULE=mockgo-grpc
make tidy MOCKGO_MODULE=mockgo-grpc
make clean MOCKGO_MODULE=mockgo-grpc
make cover MOCKGO_MODULE=mockgo-grpc
make acctest MOCKGO_MODULE=mockgo-grpc

echo "release test ended successfully "

# execute when release tag is supplied
if [[ ! -z $MOCKGO_RELEASE ]]; then
    echo "start release $MOCKGO_RELEASE ..."

    # login in github
    gh auth login --with-token < .github/token
    gh auth status
    gh config set prompt disabled

    make mod-release MOCKGO_MODULE=mockgo
    make dep-release MOCKGO_MODULE=mockgo-standalone
    make mod-release MOCKGO_MODULE=mockgo-standalone

    make dep-release MOCKGO_MODULE=grpc-kvstore
    make mod-release MOCKGO_MODULE=grpc-kvstore

    make dep-release MOCKGO_MODULE=grpc-matchstore
    make mod-release MOCKGO_MODULE=grpc-matchstore

    make dep-release MOCKGO_MODULE=mockgo-grpc
    make mod-release MOCKGO_MODULE=mockgo-grpc

    # create release with tgz as assets
    gh release create $MOCKGO_RELEASE mockgo-standalone/cmd/bin/* mockgo-grpc/cmd/bin/* --title "mockgo-server $MOCKGO_RELEASE" --notes "mockgo-server $MOCKGO_RELEASE"
    gh auth logout -h github.com

    # push to dockerhub
    export MOCKGO_IMAGE_REGISTRY=docker.io
    docker login $MOCKGO_IMAGE_REGISTRY
    make pushdocker MOCKGO_MODULE=mockgo-standalone
    make pushdocker MOCKGO_MODULE=mockgo-grpc
fi