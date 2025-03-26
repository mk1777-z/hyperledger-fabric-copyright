package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/dgrijalva/jwt-go"
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

	// 获取当前时间作为注册时间
	currentTime := time.Now()

	// 插入新用户，同时记录注册时间
	if _, err := db.Exec("INSERT INTO user (username, password, registration_time, last_active_time) VALUES (?, ?, ?, ?)",
		user.Username, user.Password, currentTime, currentTime); err != nil {
		log.Printf("Error inserting new user: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 初始化区块链账户（初始余额设为 0）
	initAccount(conf.FundsContract, user.Username, 0.0) // 确保 conf.FundsContract 已正确初始化

	log.Printf("User %s registered successfully", user.Username)
	c.Status(consts.StatusOK)
}

// login 函数
func Login(_ context.Context, c *app.RequestContext) {
	var user conf.User
	if err := c.Bind(&user); err != nil {
		log.Printf("Error binding user data: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// 查询用户
	var storedPassword string
	err = db.QueryRow("SELECT password FROM user WHERE username = ?", user.Username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Username %s not found", user.Username)
			c.Status(http.StatusUnauthorized) // 用户不存在
		} else {
			log.Printf("Error querying user data: %v", err)
			c.Status(http.StatusInternalServerError) // 其他错误
		}
		return
	}

	if user.Password != storedPassword {
		log.Printf("Incorrect password for username %s", user.Username)
		c.Status(http.StatusUnauthorized) // 密码不正确
		return
	}

	// 更新用户最后活跃时间
	currentTime := time.Now()
	_, updateErr := db.Exec("UPDATE user SET last_active_time = ? WHERE username = ?", currentTime, user.Username)
	if updateErr != nil {
		log.Printf("Error updating last_active_time for user %s: %v", user.Username, updateErr)
		// 不要因为这个错误中断登录流程
	} else {
		log.Printf("Updated last_active_time for user %s to %v", user.Username, currentTime)
	}

	// 生成 JWT token
	expirationTime := time.Now().Add(24 * time.Hour) // 设置 token 过期时间
	claims := conf.UserClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// 使用密钥签名并生成 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(conf.Con.Jwtkey)
	if err != nil {
		log.Printf("Error generating token for user %s: %v", user.Username, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 返回 token
	c.JSON(consts.StatusOK, utils.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}
