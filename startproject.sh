#!/bin/bash

# 查看占用端口 8888 的进程
PROCESS=$(lsof -i:8888)

if [ -n "$PROCESS" ]; then
    echo "发现进程占用端口 8888：$PROCESS，正在终止..."
    
    # 强制杀掉占用端口的进程
    kill -9 $PROCESS
    echo "已清理端口 8888 的进程。"
fi

go run hyperledger-fabric-copyright