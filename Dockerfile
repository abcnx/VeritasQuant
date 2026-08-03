# FinvQuant 服务端镜像（Go 1.25.3 + Gin）：多阶段构建
# 默认端口 16001

# ---- 构建阶段 ----
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# 先复制依赖清单，利用层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo 0.1.0) -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo dev)" \
    -o /finvquant-server ./cmd/server

# ---- 运行阶段 ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S finvquant && adduser -S finvquant -G finvquant

COPY --from=builder /finvquant-server /usr/local/bin/finvquant-server

USER finvquant
EXPOSE 16001

ENTRYPOINT ["finvquant-server"]
