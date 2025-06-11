package conf

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// DB 是全局数据库连接
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 从环境变量获取数据库配置，如果没有则使用默认值
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "Fabric@2024")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "fabric")

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 打开数据库连接
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		return err
	}

	// 测试数据库连接
	err = DB.Ping()
	if err != nil {
		log.Printf("数据库Ping失败: %v", err)
		return err
	}

	// **确保 chat_messages 表存在**
	if err := ensureChatTables(); err != nil {
		// 这里不直接返回错误，避免影响主流程，但打印日志提示
		log.Printf("数据库初始化成功，但创建 chat_messages 表失败: %v", err)
	}

	// 设置连接池参数
	DB.SetMaxIdleConns(10)
	DB.SetMaxOpenConns(100)

	log.Println("数据库连接成功")
	return nil
}

// 从环境变量获取值，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ensureChatTables 创建站内信(chat_messages)表（如果不存在）。
// 该函数在 InitDB 成功连接数据库后调用，以保证站内信功能依赖的数据表就绪。
func ensureChatTables() error {
	createChatMessagesSQL := `CREATE TABLE IF NOT EXISTS chat_messages (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		conversation_id VARCHAR(255) NOT NULL,
		sender_user_id VARCHAR(255) NOT NULL,
		receiver_user_id VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_read BOOLEAN NOT NULL DEFAULT FALSE,
		INDEX idx_conversation_id (conversation_id),
		INDEX idx_sender_user_id (sender_user_id),
		INDEX idx_receiver_user_id (receiver_user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	if _, err := DB.Exec(createChatMessagesSQL); err != nil {
		log.Printf("创建 chat_messages 表失败: %v", err)
		return err
	}
	return nil
}
