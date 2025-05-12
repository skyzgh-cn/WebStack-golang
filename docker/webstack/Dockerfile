# 使用官方Go镜像作为构建环境
FROM golang:1.24.2-alpine AS builder

# 设置工作目录
WORKDIR /build

# 首先复制依赖文件以利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o webstack main.go

# 使用更小的镜像作为运行环境
FROM alpine:latest

# 更换为阿里云源并安装依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata && \
    mkdir -p /app/upload

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和所需资源
COPY --from=builder /build/webstack /app/
COPY --from=builder /build/default.sql /app/
COPY --from=builder /build/config.json /app/



# 设置时区
ENV TZ=Asia/Shanghai

# 容器启动命令
CMD ["/app/webstack"]