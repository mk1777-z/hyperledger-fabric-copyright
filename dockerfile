# 使用 Ubuntu 22.04 作为基础镜像
FROM ubuntu:22.04

# 环境变量设置
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=$PATH:/usr/local/go/bin:/home/hyperledger-fabric-copyright/bin
ENV GOPROXY=https://goproxy.cn,direct

# 使用国内镜像源提升 apt 和 go 下载速度
RUN sed -i 's|http://archive.ubuntu.com/ubuntu/|http://mirrors.aliyun.com/ubuntu/|g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y \
    curl wget vim net-tools build-essential mysql-server \
    software-properties-common ca-certificates jq

# 安装 Go 1.22.7
RUN curl -fsSL https://mirrors.aliyun.com/golang/go1.22.7.linux-amd64.tar.gz -o /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz

# 创建项目根目录
WORKDIR /home

# 拷贝整个项目（包含 hyperledger-fabric-copyright 和 sample）
COPY home /home

# 设置权限
RUN chmod +x /home/hyperledger-fabric-copyright/start.sh && \
    chmod 644 /home/hyperledger-fabric-copyright/db/init.sql

# 设置项目工作目录
WORKDIR /home/hyperledger-fabric-copyright

# 预下载依赖（避免容器启动时慢）
RUN go mod tidy

# 容器启动时执行的命令
CMD bash -c "\
    echo '🎯 启动 MySQL 服务...' && \
    service mysql start && \
    echo '⏳ 等待 MySQL 启动完成...' && sleep 5 && \
    echo '⚙️  创建数据库 fabric（如未存在）...' && \
    mysql -uroot -e 'CREATE DATABASE IF NOT EXISTS fabric;' && \
    echo '🛠️  初始化数据库数据...' && \
    mysql -uroot fabric < ./db/init.sql && \
    echo '🚀 执行 start.sh 和 go run main.go...' && \
    ./start.sh && \
    go run main.go"
