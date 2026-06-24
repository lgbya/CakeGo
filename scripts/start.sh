#!/bin/bash

# 判断传入参数是否为2个
if [ $# -ne 2 ]; then
    echo "使用方式：$0  参数1 参数2"
    echo "示例：$0 1 1"
    exit 1
fi


# 接收两个入参
platID="$1"
serverID="$2"
GAME_DIR="./bin/game_${platID}_${serverID}"

go run cmd/genrpc/main.go
sh ./scripts/proto.sh
go build -o ./${GAME_DIR}/game_server_${platID}_${serverID} ./cmd/server

#go run cmd/server/main.go
cd ${GAME_DIR}
./game_server_${platID}_${serverID}