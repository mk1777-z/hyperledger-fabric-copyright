package main

import (
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.Println("正在初始化区块链和配置...")
	// 初始化配置，这会同时初始化区块链连接
	conf.Init()

	log.Println("连接到数据库...")
	// 连接到数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s",
		conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 查询所有用户
	log.Println("查询所有用户...")
	rows, err := db.Query("SELECT username FROM user")
	if err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}
	defer rows.Close()

	var successCount, failCount int

	// 为每个用户初始化账户，初始余额设为100
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			log.Printf("扫描用户名失败: %v", err)
			failCount++
			continue
		}

		// 尝试初始化用户账户
		log.Printf("正在为用户 %s 初始化账户...", username)
		_, err := conf.FundsContract.SubmitTransaction("CreateAsset", username, "0.0")
		if err != nil {
			// 如果用户已存在，则进行充值操作
			log.Printf("用户 %s 可能已存在，尝试充值操作", username)
			_, err = conf.FundsContract.SubmitTransaction("TopUp", username, "0.0")
			if err != nil {
				log.Printf("为用户 %s 充值失败: %v", username, err)
				failCount++
			} else {
				log.Printf("成功为用户 %s 充值，金额: 0.0", username)
				successCount++
			}
		} else {
			log.Printf("成功为用户 %s 初始化账户，初始余额: 0.0", username)
			successCount++
		}
	}

	log.Printf("用户账户初始化完成！成功: %d, 失败: %d", successCount, failCount)
}
