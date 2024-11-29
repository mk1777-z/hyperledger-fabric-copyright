package mv

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

// JWT 密钥
var jwtKey = []byte("123")

// UserClaims 定义用于 JWT 的声明
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// ParseToken 验证并解析 JWT Token
func ParseToken(tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// MiddlewareFunc JWT 验证中间件
func MiddlewareFunc() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tokenString := c.GetHeader("Authorization")
		if len(tokenString) > 7 && string(tokenString[:7]) == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := ParseToken(string(tokenString))
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		// 将用户名保存到上下文，供后续使用
		c.Set("username", claims.Username)
	}
}
