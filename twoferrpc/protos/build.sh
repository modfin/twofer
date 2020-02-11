#!/usr/bin/env bash

docker run -v $(pwd)/../:/go/src/github.com/twofer/twoferrpc \
    -e "UID=$(id -u ${USER})" \
    grpc/go \
    sh -c 'cd /go/src/github.com/twofer/twoferrpc/protos && protoc --proto_path=. --go_out=plugins=grpc:.. ./*.proto && chown -R $UID ..'

