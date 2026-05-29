# 请在根目录构建镜像
FROM golang:1.25.0 AS build

ARG SERVICE_PATH=internal/iam-auth-server/v1
ARG PKG_PATH=internal/pkg
ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_TIME=unknown

WORKDIR /go/github.com/sis-shen/sup-iam
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum .
RUN go mod download

COPY internal/iam-api-server/v1/model ./internal/iam-api-server/v1/model

WORKDIR /go/github.com/sis-shen/sup-iam/${PKG_PATH}

COPY ${PKG_PATH}/keys ./keys
COPY ${PKG_PATH}/log ./log
COPY ${PKG_PATH}/proto ./proto


WORKDIR /go/github.com/sis-shen/sup-iam/${SERVICE_PATH}

COPY ${SERVICE_PATH}/rpc ./rpc
COPY ${SERVICE_PATH}/service ./service
COPY ${SERVICE_PATH}/config ./config
COPY ${SERVICE_PATH}/main.go .
COPY ${SERVICE_PATH}/go ./go


RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "-X main.Version=${VERSION} \
              -X main.CommitHash=${COMMIT_HASH} \
              -X main.BuildTime=${BUILD_TIME}" \
        -mod=readonly -o /tmp/iam-auth-server .

FROM scratch AS runtime
ENV GIN_MODE=release
COPY --from=build /tmp/iam-auth-server /app/iamauthserver
EXPOSE 8080/tcp
ENTRYPOINT ["/app/iamauthserver"]
