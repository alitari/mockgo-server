#!/bin/bash

tag=$1

set -e

mkdir -p tmp
git clone --depth 1 --branch "$tag" https://github.com/alitari/mockgo-server.git tmp
cd tmp
helm dependency build deployments/helm/mockgo-server
helm package deployments/helm/mockgo-server
cd -
mv tmp/mockgo-server-*.tgz .
rm -rf tmp
helm repo index --url https://alitari.github.io/mockgo-server .




