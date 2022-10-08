#!/bin/bash

for module in mockgo grpc-kvstore grpc-matchstore mockgo-standalone mockgo-grpc
do
    cd $module
    go test -coverprofile cover.out ./...
    cd -
done

echo -e "\nCoverage Details:\n================="

for module in mockgo grpc-kvstore grpc-matchstore mockgo-standalone mockgo-grpc
do
    echo -e "\n$module :"
    go tool cover -func=${module}/cover.out
done