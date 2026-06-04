# 请在根目录构建镜像
FROM golang:1.25.0 AS build

ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_TIME=unknown

WORKDIR /go/github.com/sis-shen/sup-iam
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum .
RUN go mod download

COPY internal/iam-api-server/v1 ./internal/iam-api-server/v1
COPY internal/pkg ./internal/pkg

WORKDIR /go/github.com/sis-shen/sup-iam/internal/iam-api-server/v1

RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "-X main.Version=${VERSION} \
              -X main.CommitHash=${COMMIT_HASH} \
              -X main.BuildTime=${BUILD_TIME}" \
        -mod=readonly -o /tmp/iam-api-server .

FROM scratch AS runtime
ENV GIN_MODE=release
COPY --from=build /tmp/iam-api-server /app/iamapiserver
EXPOSE 8080/tcp
ENTRYPOINT ["/app/iamapiserver"]
