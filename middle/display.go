package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

// UserClaims 用于 JWT 的声明
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func Display(_ context.Context, c *app.RequestContext) {
	// 获取 Authorization header
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}
	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 返回 JWT 密钥
		return []byte(conf.Con.Jwtkey), nil
	})
	if err != nil {
		// 如果 token 无效，返回 401 未授权错误
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 Token 是否有效
	if _, ok := token.Claims.(*UserClaims); !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}
	defer db.Close() // 确保数据库连接在结束时关闭

	// 查询数据库，获取该用户的项目列表
	rows, err := db.Query("SELECT id, name, simple_dsc, owner , price,img FROM item WHERE on_sale=1 ORDER BY RAND() LIMIT 20 ")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database query error"})
		return
	}
	defer rows.Close()

	// 存储查询结果
	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var name, simple_des string
		var price float32
		var owner string
		var img string // 使用 sql.NullString 来处理可能的 NULL 值

		// 扫描数据
		if err := rows.Scan(&id, &name, &simple_des, &owner, &price, &img); err != nil {
			c.Status(http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, utils.H{"message": "Error reading row"})
			return
		}

		if img == "NULL" {
			img = "noimage"
		}

		// 将结果添加到 items
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": simple_des,
			"price":       price,
			"img":         img,
			"owner":       owner,
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"message": "Items fetched successfully",
		"items":   items,
	})
}
