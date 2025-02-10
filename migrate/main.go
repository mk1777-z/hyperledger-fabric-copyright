package main

import (
	"fmt"
	"log"
	"time"

	"database/sql"
	"hyperledger-fabric-copyright/conf"

	_ "github.com/go-sql-driver/mysql" // 导入 MySQL 驱动程序

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func main() {
	conf.Init()

	// 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s",
		conf.Con.Mysql.DbUser,
		conf.Con.Mysql.DbPassword,
		conf.Con.Mysql.DbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 初始化Fabric客户端
	conf.Init()

	// 查询需要初始化的版权记录
	rows, err := db.Query(`
		SELECT name, owner, start_time, price 
		FROM item 
		WHERE transID IS NULL OR transID = ''
	`)
	if err != nil {
		log.Fatal("数据库查询失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name      string
			owner     string
			startTime string
			price     float64
		)

		if err := rows.Scan(&name, &owner, &startTime, &price); err != nil {
			log.Printf("数据解析失败: %v", err)
			continue
		}

		// 解析时间字符串（格式为 "2006-01-02"）
		t, err := time.Parse("2006-01-02", startTime)
		if err != nil {
			log.Printf("时间格式错误: %s", startTime)
			continue
		}

		// 生成唯一assetID（添加毫秒时间戳+名称哈希）
		timestamp := t.UnixNano() / 1e6
		//assetID := fmt.Sprintf("asset%d-%s", timestamp, name)
		assetID := fmt.Sprintf("asset%d", timestamp)

		// 创建初始交易记录
		createInitialTransaction(conf.BasicContract, name, assetID, owner, t)

		// 更新数据库
		if _, err := db.Exec(
			"UPDATE item SET transID = ? WHERE name = ?",
			assetID,
			name,
		); err != nil {
			log.Printf("数据库更新失败: %v", err)
		}
	}
}

// 新增：带重试机制的合约调用
func createInitialTransaction(contract *client.Contract, name string, assetID string, owner string, t time.Time) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		_, err := contract.SubmitTransaction("CreateCreatetrans",
			assetID,
			name,
			"admin",                         // 固定为admin
			owner,                           // 当前所有者
			"0",                             // 初始价格为0
			t.Format("2006-01-02 15:04:05"), // 格式化时间为 "YYYY-MM-DD 00:00:00"
		)

		if err == nil {
			log.Printf("已创建初始交易: %s", assetID)
			return
		}

		log.Printf("链码调用失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second) // 指数退避更佳
	}
	log.Printf("创建交易失败: %s", assetID)
}
