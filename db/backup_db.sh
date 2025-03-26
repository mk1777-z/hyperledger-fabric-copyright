#!/bin/bash

# 数据库连接信息 - 请修改这些值以匹配您的环境
DB_USER="root"
DB_PASS="Fabric@2024"
DB_NAME="fabric"
BACKUP_DIR="/home/hyperledger-fabric-copyright/db/backups"

# 创建备份目录（如果不存在）
mkdir -p $BACKUP_DIR

# 生成带时间戳的文件名
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_backup_$TIMESTAMP.sql"
COMPRESSED_BACKUP_FILE="$BACKUP_FILE.gz"

# 输出开始备份的提示
echo "开始备份数据库 $DB_NAME 到 $BACKUP_FILE"

# 执行备份命令
mysqldump -u $DB_USER -p$DB_PASS --databases $DB_NAME > $BACKUP_FILE

# 检查备份是否成功
if [ $? -eq 0 ]; then
    echo "备份成功创建"
    
    # 压缩备份文件以节省空间
    echo "压缩备份文件..."
    gzip $BACKUP_FILE
    
    echo "备份完成: $COMPRESSED_BACKUP_FILE"
    echo "备份大小: $(du -h $COMPRESSED_BACKUP_FILE | cut -f1)"
else
    echo "备份失败"
    exit 1
fi

# 可选：列出所有备份并保留最近7天的备份
echo "清理旧备份..."
find $BACKUP_DIR -name "${DB_NAME}_backup_*.sql.gz" -type f -mtime +7 -delete

echo "当前所有备份:"
ls -lh $BACKUP_DIR | grep "${DB_NAME}_backup_"
