package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Register(ctx context.Context, c *app.RequestContext) {
	var user conf.User
	if err := c.Bind(&user); err != nil {
		log.Printf("Error binding user data: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 检查用户名是否存在
	rows, err := db.Query("SELECT username FROM user WHERE username = ?", user.Username)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() { // 如果存在该用户
		log.Printf("Username %s already exists", user.Username)
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Username %s already exists", user.Username),
		})
		return
	}

	// 插入新用户
	if _, err := db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", user.Username, user.Password); err != nil {
		log.Printf("Error inserting new user: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 初始化区块链账户（初始余额设为 0）
	initAccount(conf.FundsContract, user.Username, 0.0) // 确保 conf.FundsContract 已正确初始化

	log.Printf("User %s registered successfully", user.Username)
	c.Status(consts.StatusOK)
}
