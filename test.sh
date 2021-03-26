#!/usr/bin/env bash

DB_PORT=5432
API_PORT=8000
DB_PASSWORD=potato

bin=/tmp/potato

function cleanup(){
    pkill potato
    sudo docker stop fibdb >>/dev/null
    sudo docker rm fibdb >>/dev/null
}

function wait_for_port(){
    for x in {1..3}; do
        sleep 1
        printf "trying port $1: "
        echo -e '\x1dclose\x0d' | telnet localhost $1 >> /dev/null 2>>/dev/null && { echo "OK"; return; } || echo "failed"
    done

    # if it fails too many times
    echo "failed to connect to $1"
    exit
}

trap cleanup INT TERM EXIT

sudo docker run -d --name fibdb -e POSTGRES_PASSWORD=$DB_PASSWORD -p $DB_PORT:5432 postgres:13.2-alpine >>/dev/null
wait_for_port $DB_PORT

go build -o /tmp/potato || exit
DB_PASSWORD=$DB_PASSWORD DB_PORT=$DB_PORT /tmp/potato &
wait_for_port $API_PORT

./fib_tests.py


