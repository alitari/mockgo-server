#!/bin/bash

set -e

misspell -error .

for module in mockgo grpc-kvstore grpc-matchstore mockgo-standalone mockgo-grpc
do
    cd $module
    gofmt -s -w .
    go vet ./...
    ineffassign ./...
    gocyclo -ignore '.*\.pb\.go$' -over 15 .
    # golint -set_exit_status ./...
    go test -coverprofile cover-temp.out ./...
    cat cover-temp.out | grep -v ".pb.go" > cover.out
    rm cover-temp.out
    cd -
done

echo -e "\nCoverage Details:\n================="

for module in mockgo grpc-kvstore grpc-matchstore mockgo-standalone mockgo-grpc
do
    echo -e "\n$module :"
    go tool cover -func=${module}/cover.out
done