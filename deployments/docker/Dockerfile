# 多阶段构建：第一阶段编译，第二阶段运行。
# 注意：基础镜像 tag 需要随宿主 Go 版本同步升级。
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gorgi ./cmd/gorgi

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/gorgi /app/gorgi
COPY configs/config.yaml.example /app/configs/config.yaml.example

ENV GORGI_CONFIG=/app/configs/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/gorgi"]
