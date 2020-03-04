#!/usr/bin/env bash

docker run -v $(pwd)/../:/go/src/github.com/twofer/twoferrpc \
    -e "UID=$(id -u ${USER})" \
    grpc/go \
    sh -c '
    cd /go/src/github.com/twofer/twoferrpc/protos &&
    mkdir -p ../geid &&
    mkdir -p ../gotp &&
    mkdir -p ../gqr &&
    mkdir -p ../gw6n &&
    protoc --proto_path=. --go_out=plugins=grpc:../geid ./eid.proto &&
    protoc --proto_path=. --go_out=plugins=grpc:../gotp ./otp.proto &&
    protoc --proto_path=. --go_out=plugins=grpc:../gqr ./qr.proto &&
    protoc --proto_path=. --go_out=plugins=grpc:../gw6n ./w6n.proto &&
    chown -R $UID ..'

