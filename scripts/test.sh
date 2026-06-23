#!/bin/bash
set -e

if [ $# -ne 2 ]; then
    echo "使用方式：$0  平台id 区服id"
    echo "示例：$0 1 1"
    exit 1
fi
# 接收两个入参
platID="$1"
serverID="$2"
GAME_DIR="./bin/game_${platID}_${serverID}"

if [ ! -d "${GAME_DIR}" ]; then
    echo "错误：源目录 ${GAME_DIR} 不存在！"
    exit 1
fi


go build -o ./${GAME_DIR}/test_client ./cmd/test_client
cd ${GAME_DIR}
./test_client
