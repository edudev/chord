#!/bin/bash

mkdir remote
cp test-scripts/* remote/
if [[ "$OSTYPE" == "darwin"* ]]; then
  echo "Getting /x/sys/unix"
  go get golang.org/x/sys/unix
fi
protoc -I server/ server/*.proto --go_out=plugins=grpc:server
GOOS=linux GOARCH=amd64 go build github.com/edudev/chord
cp chord remote/
echo "Please copy the folder "\""remote"\"" onto DAS5 using scp"
echo "Run install.sh to download:Memcached, Memtier and Memslap"
echo "If you plan to install Go and clone remove all comments from install.sh"
