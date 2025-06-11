package middle

import (
	"context"
	"hyperledger-fabric-copyright/conf"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

// EnsureLoggedIn is a middleware to check for a valid JWT token
func EnsureLoggedIn() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		tokenString := ctx.GetHeader("Authorization")
		if string(tokenString) == "" {
			ctx.JSON(http.StatusUnauthorized, utils.H{"error": "Authorization token is missing"})
			ctx.Abort()
			return
		}

		// 提取 Bearer token
		token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

		// 解析 token
		token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
			// 返回 JWT 密钥
			return conf.Con.Jwtkey, nil
		})
		if err != nil {
			// 如果 token 无效，返回 401 未授权错误
			ctx.JSON(http.StatusUnauthorized, utils.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		// 验证 token 是否有效
		claims, ok := token.Claims.(*conf.UserClaims)
		if !ok || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, utils.H{"error": "Invalid token claims"})
			ctx.Abort()
			return
		}

		// Set user information in context
		ctx.Set("user", claims) // Storing the whole claims object
		ctx.Next(c)
	}
}
