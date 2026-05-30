# 请在项目根目录构建镜像
FROM ubuntu:22.04

ARG DEPLOY_PATH=deploy/LVS
RUN apt-get update && \
    apt-get install -y ipvsadm iproute2 curl && \
    apt-get clean

COPY ./$DEPLOY_PATH/lvs.sh /lvs.sh
RUN chmod +x /lvs.sh

# 使用 ENTRYPOINT + CMD 组合
ENTRYPOINT ["/lvs.sh"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ipvsadm -ln 2>/dev/null | grep -q "^TCP" || exit 1


LABEL description="LVS Director Container"