#!/bin/bash
set -e

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


#===========创建数据库

# SQL脚本路径
SQL_FILE="./sql/game_db.sql"
# 临时SQL文件（防止修改原模板）
TMP_SQL="./sql/tmp_game_db.sql"
# 目标数据库名
TARGET_DB="game_db_${platID}_${serverID}"

cp "${SQL_FILE}" "${TMP_SQL}"
if [[ $(uname -s) == MINGW* || $(uname -s) == MSYS* ]]; then
    sed -i '' "s/game_db/${TARGET_DB}/g" "${TMP_SQL}"
else
    sed -i "s/game_db/${TARGET_DB}/g" "${TMP_SQL}"
fi

YAML_FILE="${TARGET_ENV}/app.yaml"

MYSQL_HOST=$(get_yaml_val "host")
MYSQL_PORT=$(get_yaml_val "port")
MYSQL_USER=$(get_yaml_val "user")
MYSQL_PASS=$(get_yaml_val "pass")


# 执行SQL脚本
echo "开始执行SQL初始化库表：${TARGET_DB}"


# 执行mysql，端口不加引号
mysql -h${MYSQL_HOST} -P${MYSQL_PORT} -u${MYSQL_USER} -p${MYSQL_PASS} < "${TMP_SQL}"

# 判断SQL执行结果
if [ $? -eq 0 ]; then
    echo "数据库 ${TARGET_DB} 初始化成功！"
    # 删除临时SQL文件
    rm -f "${TMP_SQL}"
else
    echo "数据库初始化失败，请检查MySQL服务或账号密码！"
    rm -f "${TMP_SQL}"
    exit 1
fi

go run cmd/genrpc/main.go
sh ./scripts/proto.sh
#go run cmd/server/main.go
go build -o ./${GAME_DIR}/game_server_${platID}_${serverID} ./cmd/server

echo "===== 部署完成 ====="


