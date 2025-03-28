package middle

import (
	"context"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

// GetBalance 处理获取余额的请求
func GetBalance(_ context.Context, c *app.RequestContext) {
	fmt.Println("处理获取余额请求...")

	// 从请求中获取用户名
	var req struct {
		Username string `json:"username"`
	}

	if err := c.Bind(&req); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body"})
		return
	}

	// 验证token - 修复类型问题
	tokenString := string(c.GetHeader("Authorization"))
	if tokenString == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取Bearer token
	token_String := strings.Replace(tokenString, "Bearer ", "", -1)

	// 解析token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}
	username := claims.Username

	// 确保查询的是自己的余额
	if req.Username != username {
		c.Status(http.StatusForbidden)
		c.JSON(http.StatusForbidden, utils.H{"message": "Cannot query balance of other users"})
		return
	}

	// 调用账户余额查询函数
	balance, err := queryBalance(conf.FundsContract, username)
	if err != nil {
		// 如果账户不存在，返回0余额
		if strings.Contains(err.Error(), "does not exist") {
			c.JSON(http.StatusOK, utils.H{
				"balance": 0.0,
			})
			return
		}
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": err.Error()})
		return
	}

	// 返回查询结果
	c.JSON(http.StatusOK, utils.H{
		"balance": balance,
	})
}
