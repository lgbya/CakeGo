#!/bin/bash
set -e

# 判断传入参数是否为2个
if [ $# -ne 2 ]; then
    echo "使用方式：$0  参数1 参数2"
    echo "示例：$0 1 1"
    exit 1
fi

platID="$1"
serverID="$2"
APP_NAME="game_server_${platID}_${serverID}"
echo "待关闭进程名：${APP_NAME}"

# -f 匹配完整命令行，适配长进程名
PID=$(pgrep -f "${APP_NAME}")

if [ -n "$PID" ];then
    echo "正在关闭进程 PID: $PID"
    # 优雅信号 SIGTERM
    kill -15 $PID

    # 等待10秒优雅退出
    wait_time=600
    for ((i=0; i<wait_time; i++)); do
        if ! pgrep -f "${APP_NAME}" >/dev/null; then
            echo "进程已正常退出"
            exit 0
        fi
        sleep 1
    done

    # 超时未退出，强制杀掉
    echo "关闭超时 ${wait_time}s，强制杀死进程"
    kill -9 $PID
else
    echo "进程【${APP_NAME}】未运行"
fi