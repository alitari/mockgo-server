#!/bin/bash

set -e

cd grpc-matchstore
rm -f matchstore/*.pb.go  
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative matchstore/matchstore.proto
echo "generated files:"
ls -l ./matchstore/*.pb.go
cd -
