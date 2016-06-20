#!/bin/bash

go version 1>/dev/null
if [ $? -ne 0 ]; then
    echo "Golang is not installed or you must add go to your PATH"
    exit 1
fi

echo "Installing Golang dependencies..."
go get golang.org/x/net/html
go get gopkg.in/redis.v3
go get patrickmn/go-cache

echo "Building crawler"
go build -o crawler ./src

if [ $? -eq 0 ]; then
    echo "Build Successful"
    exit 0
else
    echo "Build Unsuccessful"
    exit 1
fi
