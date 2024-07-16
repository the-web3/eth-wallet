#!/bin/bash

function exit_if() {
    extcode=$1
    msg=$2
    if [ $extcode -ne 0 ]
    then
        if [ "msg$msg" != "msg" ]; then
            echo $msg >&2
        fi
        exit $extcode
    fi
}

echo $GOPATH;

if [ ! -f $GOPATH/bin/protoc-gen-go ]
then
    echo 'No plugin for golang installed, skip the go installation' >&2
    echo 'try go get github.com/golang/protobuf/protoc-gen-go' >&2
else
    echo Compiling go interfaces...
    export GO_PATH=$GOPATH
    export GOBIN=$GOPATH/bin
    export PATH=$PATH:$GOPATH/bin

    protoc -I ./ --go_out=./ --go-grpc_out=./ rpc/*.proto

    exit_if $?
    echo Done
fi