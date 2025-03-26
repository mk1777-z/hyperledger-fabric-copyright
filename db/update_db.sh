#!/bin/bash

# 请替换为您的MySQL连接信息
MYSQL_USER="your_user"
MYSQL_PASS="your_password"
MYSQL_DB="your_database"

echo "开始更新数据库结构..."
mysql -u $MYSQL_USER -p$MYSQL_PASS $MYSQL_DB < schema_update.sql

echo "开始填充随机数据..."
mysql -u $MYSQL_USER -p$MYSQL_PASS $MYSQL_DB < data_fill.sql

echo "数据库更新完成！"
