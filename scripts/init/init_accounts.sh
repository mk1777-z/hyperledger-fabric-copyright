#!/bin/bash

echo "开始初始化所有用户的账户..."
# 从项目根目录运行，确保能正确导入包
cd /home/hyperledger-fabric-copyright
go run scripts/init/init_accounts.go
echo "初始化完成！"
