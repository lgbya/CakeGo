#!/bin/bash

trim() {
    echo "$1" | sed 's/\r//g' | xargs
}

# 从yaml读取指定key的值
get_yaml_val() {
    local key="$1"
    local raw_val
    raw_val=$(sed -n "s/^[[:space:]]*${key}:[[:space:]]*\"\\(.*\\)\".*/\\1/p" "${YAML_FILE}")
    trim "$raw_val"
}

#==========================================
# 判断传入参数是否为2个
if [ $# -ne 2 ]; then
    echo "使用方式：$0  参数1 参数2"
    echo "示例：$0 1 1"
    exit 1
fi
if [ $# -ne 2 ]; then
    echo "使用方式：$0  参数1 参数2"
    echo "示例：$0 1 1"
    exit 1
fi

# 接收两个入参
platID="$1"
serverID="$2"

# 目标根目录
GAME_DIR="./bin/game_${platID}_${serverID}"
# 目标env目录
TARGET_ENV="${GAME_DIR}/env"
# 源env目录（当前目录下的env）
SOURCE_CONFIG="./env"

# 1. 创建游戏根目录
if [ -d "${GAME_DIR}" ]; then
    echo "错误：源目录 ${GAME_DIR} 已存在！"
    exit 1
fi
mkdir -p "${GAME_DIR}"
echo "游戏目录准备完成: ${GAME_DIR}"

# 2. 创建目标config目录
mkdir -p "${TARGET_ENV}"
echo "目标config目录: ${TARGET_ENV}"

# 3. 判断源config是否存在
if [ ! -d "${SOURCE_CONFIG}" ]; then
    echo "错误：源目录 ${SOURCE_CONFIG} 不存在，无法复制yaml文件！"
    exit 1
fi

# 4. 复制所有 .yaml .yml 文件
if ls "${SOURCE_CONFIG}"/*.yaml >/dev/null 2>&1; then
    cp "${SOURCE_CONFIG}"/*.yaml "${TARGET_ENV}/"
    echo "已复制所有 .yaml 文件到 ${TARGET_ENV}"
fi

if ls "${SOURCE_CONFIG}"/*.yml >/dev/null 2>&1; then
    cp "${SOURCE_CONFIG}"/*.yml "${TARGET_ENV}/"
    echo "已复制所有 .yml 文件到 ${TARGET_ENV}"
fi
sed -i "s/serverId: [0-9]*/serverId: ${serverID}/" "${TARGET_ENV}/app.yaml"
sed -i "s/platId : [0-9]*/platId : ${platID}/" "${TARGET_ENV}/app.yaml"
sed -i "s/name: \"game_db\"/name: \"game_db_${platID}_${serverID}\"/" "${TARGET_ENV}/app.yaml"

#go mod tidy
go run cmd/genrpc/main.go
sh ./scripts/proto.sh
#go run cmd/server/main.go
go build -o ./${GAME_DIR}/game_server_${platID}_${serverID} ./cmd/server

echo "===== 部署完成 ====="


