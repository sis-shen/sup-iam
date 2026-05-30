#!/bin/bash

# 默认值
VIP="${VIP:-192.168.1.100}"
PORT="${PORT:-80}"
SCHEDULER="${SCHEDULER:-rr}"
MODE="${MODE:-m}"  # m=NAT, g=DR, i=TUN
INTERFACE="${INTERFACE:-auto}"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --vip)
            VIP="$2"
            shift 2
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --scheduler)
            SCHEDULER="$2"
            shift 2
            ;;
        --mode)
            MODE="$2"
            shift 2
            ;;
        --interface)
            INTERFACE="$2"
            shift 2
            ;;
        --debug)
            DEBUG=1
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --vip IP         Virtual IP address (default: 192.168.1.100)"
            echo "  --port PORT      Service port (default: 80)"
            echo "  --scheduler ALG  Scheduling algorithm: rr|wrr|lc|wlc|lblc|sh|dh (default: rr)"
            echo "  --mode MODE      LVS mode: m(NAT)|g(DR)|i(TUN) (default: m)"
            echo "  --interface NIC  Network interface (default: auto)"
            echo "  --debug          Enable debug mode"
            echo "  --help           Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage"
            exit 1
            ;;
    esac
done

# 调试模式
if [ -n "$DEBUG" ]; then
    set -x
    echo "=== Debug mode enabled ==="
fi

echo "=== Configuring LVS Director ==="
echo "VIP: $VIP:$PORT"
echo "Scheduler: $SCHEDULER"
echo "Mode: $MODE"
echo "DEBUG: $DEBUG"

# 开启内核 IP 转发
echo 1 > /proc/sys/net/ipv4/ip_forward
echo "IP forwarding enabled"

# 配置 VIP（自动检测或手动指定网卡）
if [ "$INTERFACE" = "auto" ]; then
    NIC=$(ip -o link show | grep -v lo | head -1 | awk -F': ' '{print $2}')
else
    NIC="$INTERFACE"
fi

if [ -n "$NIC" ]; then
    ip addr add ${VIP}/24 dev $NIC 2>/dev/null || {
        echo " Warning: Failed to add VIP to $NIC"
    }
    echo "VIP $VIP added to $NIC"
fi

# 清除所有已有规则
ipvsadm -C
echo "Existing rules cleared"

# 添加 LVS 服务
# 根据模式选择参数：-m (NAT), -g (DR), -i (TUN)
MODE_PARAM="-m"
case $MODE in
    m|nat|NAT)  MODE_PARAM="-m" ;;
    g|dr|DR)    MODE_PARAM="-g" ;;
    i|tun|TUN)  MODE_PARAM="-i" ;;
esac

ipvsadm -A -t ${VIP}:${PORT} -s $SCHEDULER
echo "LVS service added: ${VIP}:${PORT} ($SCHEDULER)"

# 添加后端真实服务器
if [ -n "$REAL_SERVERS" ]; then
    IFS=',' read -ra SERVERS <<< "$REAL_SERVERS"
    for server in "${SERVERS[@]}"; do
        ipvsadm -a -t ${VIP}:${PORT} -r $server $MODE_PARAM
        echo "Real server added: $server"
    done
elif [ $# -eq 0 ]; then
    echo "ℹ No real servers configured"
    echo "  Set REAL_SERVERS env var or use command line:"
    echo "  docker run -e REAL_SERVERS=\"IP1:PORT,IP2:PORT\" ..."
fi

# 打印当前配置
echo ""
echo "=== Current LVS Configuration ==="
ipvsadm -ln
echo ""

# 优雅退出处理
trap 'echo "Shutting down LVS..."; ipvsadm -C; echo "Done"; exit 0' SIGTERM SIGINT

# 保持容器运行
echo "=== LVS Director is running ==="
echo "Configuration: VIP=$VIP:$PORT, Scheduler=$SCHEDULER, Mode=$MODE"
while true; do
    sleep 3600
done