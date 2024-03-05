#!/bin/bash

## DIRECT
 # ie. ./ll-go-run.sh serve ./file.go arg1 ... argN
## DOCKER
 # CMD ["./ll-go-run.sh", "serve", "./file.go", "arg1", "...", "argN"]

## Dependencies
 # inotify-hookable, sudo apt-get install inotify-hookable

SCRIPT=$0
CMD=$1
SUB_CMD=$2
FILENAME=$3
ARGS=$(echo $@ | cut -d " " -f3-)

PID_FILE=/tmp/livereload-go-run.pid

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

case $CMD in

    start)
            go run -race $FILENAME $ARGS &
            echo $! > $PID_FILE
            printf "${RED}⚡ ${GREEN}Restarting Service\n"

    ;;

    serve)
            $SCRIPT $SUB_CMD $FILENAME $ARGS
            printf "${RED}⚡ ${GREEN}Serving with live reload${NC}\n"
            ## Watches All files in dir
             # inotify-hookable -q -r -w . -c "$SCRIPT reload $FILENAME $ARGS"

            ## Watches all go file present in repo on init
            #  inotify-hookable -q -f $(find . -name "*.go" | paste -sd" " | sed 's/ / -f /g')\
            #  -c "$SCRIPT reload $FILENAME $ARGS"

            ## Watches all go in root dir along with all non hidden dir
            inotify-hookable -q \
                -f $(find ./cmd -name "*.go" | paste -sd" " | sed 's/ / -f /g') \
                -r -w internal \
                -f go.mod \
                -c "$SCRIPT reload $SUB_CMD $FILENAME $ARGS"
    ;;
    reload)

            LT=$(stat ${PID_FILE} | grep Change | cut -d ' ' -f2,3,4 | cut -d '.' -f1)
            DEB=$(date '+%F %T' -d '-4 seconds')
            echo "$DEB < $LT"
            if [[ ${DEB} < ${LT} ]]; then
                echo debounce
                exit 0
            fi
            touch $PID_FILE

            pkill -P $(cat $PID_FILE)
            sleep 1

            printf "${RED}⚡ ${GREEN}Reloading Service with ${SUB_CMD} ${NC}\n"

            $SCRIPT $SUB_CMD $FILENAME $ARGS
    ;;
esac
