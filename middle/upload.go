package middle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

func Upload(_ context.Context, c *app.RequestContext) {
	var uploadInfo conf.Upload
	c.Bind(&uploadInfo)
	// 获取 Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 验证 Authorization 头部格式是否正确
	if !strings.HasPrefix(string(tokenString), "Bearer ") {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid Authorization header format"})
		return
	}

	// 提取 Bearer Token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 Token
	token, err := jwt.ParseWithClaims(token_String, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 验证签名方法是否匹配
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(conf.Con.Jwtkey), nil
	})
	if err != nil {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				c.Status(http.StatusUnauthorized)
				c.JSON(http.StatusUnauthorized, utils.H{"message": "Token has expired"})
				return
			}
		}
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 Token 的 Claims
	if !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		log.Fatal("OPEN SQL ERROR")
		return
	}
	defer db.Close()

	exists, _ := db.Query("SELECT * FROM item WHERE name = ? AND id = ? ", uploadInfo.Name, uploadInfo.ID)
	if exists.Next() {
		c.Status(http.StatusConflict)
		c.JSON(http.StatusConflict, utils.H{"message": "Item Already Exist"})
		return
	}

	_, err = db.Exec(
		"INSERT INTO item (id, name, owner, simple_dsc, dsc, price, img, on_sale, start_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uploadInfo.ID,
		uploadInfo.Name,
		token.Claims.(*UserClaims).Username,
		uploadInfo.Simple_dsc,
		uploadInfo.Dsc,
		uploadInfo.Price,
		uploadInfo.Img,
		uploadInfo.On_sale,
		time.Now(),
	)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Internal Server Error"})
		log.Fatal(err)
		return
	}
	c.Status(http.StatusOK)
	c.JSON(http.StatusOK, utils.H{"message": "Create item success"})
}
