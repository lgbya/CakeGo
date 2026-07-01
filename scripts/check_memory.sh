#!/bin/bash
# scripts/check_memory.sh
# 检查游戏服务器堆内存

# 检查 pprof 是否可访问
if ! curl -s http://localhost:6060/debug/pprof/ > /dev/null 2>&1; then
    echo "⚠️ pprof 未启用（http://localhost:6060 不可访问）"
    exit 0
fi

echo "📊 正在采集内存数据..."

# 获取 HeapInuse 原始行，例如: "# HeapInuse: 8904704"
RAW_LINE=$(curl -s http://localhost:6060/debug/pprof/heap?debug=1 | grep "# HeapInuse" | head -1)

if [ -z "$RAW_LINE" ]; then
    echo "❌ 无法解析内存数据"
    exit 1
fi

# 从行中提取纯数字（去掉空格、冒号、等号等非数字字符）
INUSE_BYTES=$(echo "$RAW_LINE" | grep -oE '[0-9]+' | head -1)

# 如果提取到的数字为空或不是纯数字，报错
if [ -z "$INUSE_BYTES" ] || ! echo "$INUSE_BYTES" | grep -qE '^[0-9]+$'; then
    echo "❌ 内存数据格式异常: '$RAW_LINE'"
    exit 1
fi

# 用 awk 计算 MB（保留1位小数）
TOTAL_MEM=$(awk "BEGIN {printf \"%.1f\", $INUSE_BYTES / 1024 / 1024}")

# 阈值 50MB
THRESHOLD=50

echo "📊 当前存活内存: ${TOTAL_MEM}MB"

# 取整比较（去掉小数）
TOTAL_MEM_INT=$(echo "$TOTAL_MEM" | awk '{print int($1)}')

if [ "$TOTAL_MEM_INT" -gt "$THRESHOLD" ] 2>/dev/null; then
    echo "❌ 内存超标: ${TOTAL_MEM}MB > ${THRESHOLD}MB"
    exit 1
else
    echo "✅ 内存正常"
    exit 0
fi