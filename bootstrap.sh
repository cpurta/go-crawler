#!/bin/bash

go version
if [ $? -ne 0 ]; then
    echo "Golang is not installed or you must add go to your PATH"
    exit 1
fi

echo "Installing Golang dependencies..."
go get golang.org/x/net/html

exit 0
