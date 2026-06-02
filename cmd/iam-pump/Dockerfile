# 请在根目录构建镜像
FROM golang:1.25.0 AS build

ARG APP_PATH=internal/iam-pump
ARG PKG_PATH=internal/pkg
ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_TIME=unknown

WORKDIR /go/github.com/sis-shen/sup-iam
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum .
RUN go mod download

COPY cmd/iam-pump/main.go ./cmd/iam-pump/main.go

WORKDIR /go/github.com/sis-shen/sup-iam/${PKG_PATH}

COPY ${PKG_PATH}/log ./log
COPY ${PKG_PATH}/testutil ./testutil
COPY ${PKG_PATH}/server ./server
COPY ${PKG_PATH}/options ./options


WORKDIR /go/github.com/sis-shen/sup-iam/${APP_PATH}

COPY ${APP_PATH}/analytics ./analytics
COPY ${APP_PATH}/store ./store
COPY ${APP_PATH}/options ./options
COPY ${APP_PATH}/pumps ./pumps
COPY ${APP_PATH}/config ./config
COPY ${APP_PATH}/app.go ./app.go
COPY ${APP_PATH}/run.go ./run.go
COPY ${APP_PATH}/server.go ./server.go

WORKDIR /go/github.com/sis-shen/sup-iam/cmd/iam-pump

RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "-X main.Version=${VERSION} \
              -X main.CommitHash=${COMMIT_HASH} \
              -X main.BuildTime=${BUILD_TIME}" \
        -mod=readonly -o /tmp/iam-pump .

FROM scratch AS runtime
ENV GIN_MODE=release
COPY --from=build /tmp/iam-pump /app/iam-pump
EXPOSE 8080/tcp
ENTRYPOINT ["/app/iam-pump"]
