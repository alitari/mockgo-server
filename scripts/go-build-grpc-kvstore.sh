#!/bin/sh

cd grpc-kvstore
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative kvstore/kvstore.proto
echo "generated files:"
ls -l ./kvstore/*.pb.go
cd -
