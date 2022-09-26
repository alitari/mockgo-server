#!/bin/sh

cd grpc-matchstore
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative matchstore/matchstore.proto
echo "generated files:"
ls -l ./matchstore/*.pb.go
cd -
