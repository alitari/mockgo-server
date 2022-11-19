#!/bin/bash

if [ $# -eq 0 ]
then
    echo "A release tag must be supplied"
    exit 1
fi
releaseTag=$1

gitsha=$(git rev-parse --short HEAD)

set -e

PATH="$PATH:$(go env GOPATH)/bin" 

# git checks
branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$branch" != "master" ]
then
    echo "you must be in 'master' branch, but you are in '$branch'!"
    exit 1
fi
gitstatus=$(git status --short)
if [[ "$gitstatus" != "" ]]; then
    echo "the workspace is dirty: $gitstatus !"
    exit 1
fi




# publish mockgo module 
cd mockgo
go mod tidy
go clean -testcache ./...
go test ./...

mockgotag="mockgo/$releaseTag"

echo "tagging mockgo module with '$mockgotag' ..."

git tag -a $mockgotag -m "🔖 Tag mockgo module with $mockgotag"
git push origin $mockgotag

GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/mockgo@$releaseTag"
cd -

# mockgo standalone 

cd mockgo-standalone
go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"
go mod tidy
go clean -testcache ./...
go test ./...

mockgostandalonetag="mockgo-standalone/$releaseTag"
echo "tagging mockgo-standalone module with '$mockgostandalonetag' ..."
git tag -a $mockgostandalonetag -m "🔖 Tag mockgo-standalone module with $mockgostandalonetag"
git push origin $mockgostandalonetag

GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/mockgo-standalone@$releaseTag"
cd -

# grpc-kvstore
cd grpc-kvstore
go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"
go mod tidy
go clean -testcache ./...
go test ./...

grpckvstoretag="grpc-kvstore/$releaseTag"
echo "tagging grpc-kvstore module with '$grpckvstoretag' ..."
git tag -a $grpckvstoretag -m "🔖 Tag grpc-kvstore module with $grpckvstoretag"
git push origin $grpckvstoretag

GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/grpc-kvstore@$releaseTag"
cd -

# grpc-matchstore
cd grpc-matchstore
go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"
go mod tidy
go clean -testcache ./...
go test ./...

grpcmatchstoretag="grpc-matchstore/$releaseTag"
echo "tagging grpc-matchstore module with '$grpcmatchstoretag' ..."
git tag -a $grpcmatchstoretag -m "🔖 Tag grpc-matchstore module with $grpcmatchstoretag"
git push origin $grpcmatchstoretag

GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/grpc-matchstore@$releaseTag"
cd -

# mockgo-grpc
cd mockgo-grpc

go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"

go mod edit -droprequire github.com/alitari/mockgo-server/grpc-kvstore
go mod edit -dropreplace github.com/alitari/mockgo-server/grpc-kvstore
go mod edit -require "github.com/alitari/mockgo-server/grpc-kvstore@$releaseTag"

go mod edit -droprequire github.com/alitari/mockgo-server/grpc-matchstore
go mod edit -dropreplace github.com/alitari/mockgo-server/grpc-matchstore
go mod edit -require "github.com/alitari/mockgo-server/grpc-matchstore@$releaseTag"

go mod tidy
go clean -testcache ./...
go test ./...

mockgogrprctag="mockgo-grpc/$releaseTag"
echo "tagging mockgo-grpc module with '$mockgogrprctag' ..."
git tag -a $mockgogrprctag -m "🔖 Tag mockgo-grpc module with $mockgogrprctag"
git push origin $mockgogrprctag

GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/mockgo-grpc@$releaseTag"
cd -
