#!/bin/bash

# function for make clean all modules
function make_clean_all() {
    for module in mockgo mockgo-standalone grpc-kvstore grpc-matchstore mockgo-grpc
    do
        make clean MOCKGO_MODULE=$module
    done
}

# function for make mod-dev all modules
function make_dep_dev_all() {
    for module in mockgo-standalone grpc-kvstore grpc-matchstore mockgo-grpc
    do
        make dep-dev MOCKGO_MODULE=$module
    done
}

#function for make mod-release all modules
function make_dep_release_all() {
    for module in mockgo-standalone grpc-kvstore grpc-matchstore mockgo-grpc
    do
        make dep-release MOCKGO_MODULE=$module
    done
}


if [ $# -eq 0 ]
then
    echo "start release test"
    make env
    make_clean_all
    make_dep_dev_all
else
    # check for semver format
    if [[ ! $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "release tag must be in semver format, e.g. v1.0.0"
        exit 1
    fi
    echo "start release $1"
    # git stuff
    make env
    export MOCKGO_RELEASE=$1
    make_clean_all
    make_dep_release_all

fi
gitsha=$(git rev-parse --short HEAD)

# docker login

# set -e

# PATH="$PATH:$(go env GOPATH)/bin" 

# # git checks
# branch=$(git rev-parse --abbrev-ref HEAD)
# if [ "$branch" != "master" ]
# then
#     echo "you must be in 'master' branch, but you are in '$branch'!"
#     exit 1
# fi
# gitstatus=$(git status --short)
# if [[ "$gitstatus" != "" ]]; then
#     echo "the workspace is dirty: $gitstatus !"
#     exit 1
# fi

# # create a release branch
# git checkout -b "release-$releaseTag"




# show environment for all

make build MOCKGO_MODULE=mockgo
make cover MOCKGO_MODULE=mockgo
make mod-release MOCKGO_MODULE=mockgo

make cover MOCKGO_MODULE=mockgo-standalone
make hurl MOCKGO_MODULE=mockgo-standalone
make mod-release MOCKGO_MODULE=mockgo-standalone

make build MOCKGO_MODULE=grpc-kvstore
make cover MOCKGO_MODULE=grpc-kvstore
make mod-release MOCKGO_MODULE=grpc-kvstore

make build MOCKGO_MODULE=grpc-matchstore
make cover MOCKGO_MODULE=grpc-matchstore
make mod-release MOCKGO_MODULE=grpc-matchstore

make cover MOCKGO_MODULE=mockgo-grpc
make hurl MOCKGO_MODULE=mockgo-grpc
make mod-release MOCKGO_MODULE=mockgo-grpc

# for module in mockgo mockgo-standalone grpc-kvstore grpc-matchstore mockgo-grpc
# do
#     make clean MOCKGO_MODULE=$module
# done

    # # delete binaries folder
    # rm -f ./$moduls/bin/*

    # # test mockgo module 
    # cd $moduls
    # go mod tidy
    # go clean -testcache ./...
    # go test ./...

    # # publish mockgo
    # mockgotag="$moduls/$releaseTag"
    # echo "tagging $moduls module with '$mockgotag' ..."
    # git tag -a $mockgotag -m "🔖 Tag $moduls module with $mockgotag"
    # git push origin $mockgotag
    # GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/$moduls@$releaseTag"
    # cd -

# # test mockgo module 
# cd mockgo
# go mod tidy
# go clean -testcache ./...
# go test ./...

# # publish mockgo
# mockgotag="mockgo/$releaseTag"
# echo "tagging mockgo module with '$mockgotag' ..."
# git tag -a $mockgotag -m "🔖 Tag mockgo module with $mockgotag"
# git push origin $mockgotag
# GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/mockgo@$releaseTag"
# cd -

# # mockgo standalone setup dependencies & test
# cd mockgo-standalone
# go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
# go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
# go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"
# go mod tidy
# go clean -testcache ./...
# go test ./...
# cd -

# # mockgo standalone create executabels
# for target in amd64 arm64
# do
#     ./scripts/go-build-mockgo.sh linux $target standalone ${releaseTag}-${gitsha}
# done
# ./scripts/go-build-mockgo.sh windows amd64 standalone ${releaseTag}-${gitsha}

# git add -A
# git commit -m "🔖 Setup mockgo-standalone dependencies for release-$releaseTag"
# git push --set-upstream origin "release-$releaseTag"


# # mockgo standalone publish module
# mockgostandalonetag="mockgo-standalone/$releaseTag"
# echo "tagging mockgo-standalone module with '$mockgostandalonetag' ..."
# git tag -a $mockgostandalonetag -m "🔖 Tag mockgo-standalone module with $mockgostandalonetag"
# git push origin $mockgostandalonetag
# GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/mockgo-standalone@$releaseTag"

# # grpc-kvstore setup dependencies and test
# cd grpc-kvstore
# go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
# go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
# go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"
# cd -
# ./scripts/protoc-grpc-kvstore.sh

# cd grpc-kvstore
# go mod tidy
# go clean -testcache ./...
# go test ./...
# cd -

# git add -A
# git commit -m "🔖 Setup grpc-kvstore dependencies for release-$releaseTag"
# git push --set-upstream origin "release-$releaseTag"

# # grpc-kvstore publish module
# grpckvstoretag="grpc-kvstore/$releaseTag"
# echo "tagging grpc-kvstore module with '$grpckvstoretag' ..."
# git tag -a $grpckvstoretag -m "🔖 Tag grpc-kvstore module with $grpckvstoretag"
# git push origin $grpckvstoretag
# GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/grpc-kvstore@$releaseTag"

# # grpc-matchstore setup dependencies and test
# cd grpc-matchstore
# go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
# go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
# go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"
# cd -
# ./scripts/protoc-grpc-matchstore.sh

# cd grpc-matchstore
# go mod tidy
# go clean -testcache ./...
# go test ./...
# cd -

# git add -A
# git commit -m "🔖 Setup grpc-matchstore dependencies for release-$releaseTag"
# git push --set-upstream origin "release-$releaseTag"

# # grpc-matchstore publish module
# grpcmatchstoretag="grpc-matchstore/$releaseTag"
# echo "tagging grpc-matchstore module with '$grpcmatchstoretag' ..."
# git tag -a $grpcmatchstoretag -m "🔖 Tag grpc-matchstore module with $grpcmatchstoretag"
# git push origin $grpcmatchstoretag
# GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/grpc-matchstore@$releaseTag"

# # mockgo-grpc setup dependencies
# cd mockgo-grpc
# go mod edit -droprequire github.com/alitari/mockgo-server/mockgo
# go mod edit -dropreplace github.com/alitari/mockgo-server/mockgo
# go mod edit -require "github.com/alitari/mockgo-server/mockgo@$releaseTag"

# go mod edit -droprequire github.com/alitari/mockgo-server/grpc-kvstore
# go mod edit -dropreplace github.com/alitari/mockgo-server/grpc-kvstore
# go mod edit -require "github.com/alitari/mockgo-server/grpc-kvstore@$releaseTag"

# go mod edit -droprequire github.com/alitari/mockgo-server/grpc-matchstore
# go mod edit -dropreplace github.com/alitari/mockgo-server/grpc-matchstore
# go mod edit -require "github.com/alitari/mockgo-server/grpc-matchstore@$releaseTag"

# go mod tidy
# go clean -testcache ./...
# go test ./...
# cd -

# # mockgo-grpc create executabels
# for target in amd64 arm64
# do
# ./scripts/go-build-mockgo.sh linux $target grpc ${releaseTag}-${gitsha}
# done
# ./scripts/go-build-mockgo.sh windows amd64 grpc ${releaseTag}-${gitsha}

# git add -A
# git commit -m "🔖 Setup mockgo-grpc dependencies for release-$releaseTag"
# git push --set-upstream origin "release-$releaseTag"

# # mockgo-grpc publish modules
# mockgogrprctag="mockgo-grpc/$releaseTag"
# echo "tagging mockgo-grpc module with '$mockgogrprctag' ..."
# git tag -a $mockgogrprctag -m "🔖 Tag mockgo-grpc module with $mockgogrprctag"
# git push origin $mockgogrprctag
# GOPROXY=proxy.golang.org go list -m "github.com/alitari/mockgo-server/mockgo-grpc@$releaseTag"

# # create tgz and checksums
# for file in ./bin/*
# do
#     sha256sum ${file} > ${file}.sha256
#     tar -cvzf ${file}.tgz ${file}
#     sha256sum --check ${file}.sha256
#     rm ${file}
# done

# # login in github
# gh auth login --with-token < .github/token
# gh auth status
# gh config set prompt disabled

# # create release with tgz as assets
# gh release create $releaseTag ./bin/*.*
# gh auth logout -h github.com

# # docker builds

# ./scripts/docker-build-mockgo.sh $releaseTag standalone true

# ./scripts/docker-build-mockgo.sh $releaseTag grpc true
