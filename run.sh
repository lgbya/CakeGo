#!/bin/bash
# 帮助提示函数
show_help() {
    echo "===== 使用方式 ====="
    echo "1. 部署服务：sh run.sh install 服ID 平台ID"
    echo "   示例：sh run.sh install 1 2"
    echo "2. 启动服务：sh run.sh start 服ID 平台ID"
    echo "   示例：sh run.sh start 1 2"
    echo "3. 生成proto：sh run.sh proto"
    echo "4. 压测客户端：sh run.sh test 服ID 平台ID"

}

# 获取第一个指令参数
CMD="$1"

case "${CMD}" in
install|start|test)
    # install/start 必须携带 2个数字参数，总参数数量=3
    if [ $# -ne 3 ]; then
        echo "错误：指令【${CMD}】需要传入 服ID、平台ID 两个参数"
        show_help
        exit 1
    fi
    PlatID="$2"
    ServerID="$3"
    # 校验后两位必须是纯数字
    if ! [[ "$PlatID" =~ ^[0-9]+$ && "$ServerID" =~ ^[0-9]+$ ]]; then
        echo "错误：服ID、平台ID必须为数字"
        exit 1
    fi
    if [[ "${CMD}" == "install" ]]; then
        echo "===== 开始执行部署脚本 scripts/install.sh ${PlatID} ${ServerID} ====="
        sh ./scripts/install.sh "${PlatID}" "${ServerID}"
    elif [[ "${CMD}" == "start" ]]; then
        echo "===== 开始执行启动脚本 scripts/start.sh ${PlatID} ${ServerID} ====="
        sh ./scripts/start.sh "${PlatID}" "${ServerID}"
    elif [[ "${CMD}" == "test" ]]; then
        echo "===== 启动压测客户端 tester，服务 game_${NUM1}_${NUM2} ====="
        sh ./scripts/test.sh "${PlatID}" "${ServerID}"
    fi

    ;;

proto)
    if [ $# -ne 1 ]; then
        echo "错误：指令【proto】不需要携带任何额外参数"
        show_help
        exit 1
    fi
    echo "===== 开始执行脚本 scripts/proto.sh ====="
    sh ./scripts/proto.sh
    ;;


"")
    # 没有传任何参数
    echo "错误：未输入任何指令"
    show_help
    exit 1
    ;;

*)
    # 非法指令
    echo "错误指令！仅支持指令：install / start / proto"
    show_help
    exit 1
    ;;
esac
