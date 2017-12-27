#!/bin/sh

echo "In build.sh"
set -e

cd $(dirname $0)/..

src_root=`pwd`

dir_name=${src_root##*/}

echo "src root:" ${src_root}
echo "dir name:" ${dir_name}

mkdir -p ${GOPATH}/src

ln -s  ${src_root} ${GOPATH}/src/vespace-provisioner

cd ${GOPATH}/src/vespace-provisioner
echo "change to" `pwd`


#go build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -v -installsuffix cgo
