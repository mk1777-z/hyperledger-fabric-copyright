package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"strconv"

	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func UpdateItem(_ context.Context, c *app.RequestContext) {
	var updatedItem conf.UpdateItem
	c.Bind(&updatedItem)
	if err := c.Bind(&updatedItem); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 返回 JWT 密钥
		return conf.Con.Jwtkey, nil
	})
	if err != nil || !token.Valid {
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
	defer db.Close()

	// 更新数据库中的记录
	query := "UPDATE item SET name = ?, simple_dsc = ?, price = ?, dsc = ?, on_sale = ? WHERE id = ?"
	_, err = db.Exec(query, updatedItem.Name, updatedItem.Description, updatedItem.Price, updatedItem.Dsc, updatedItem.Sale, updatedItem.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database update error"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, utils.H{"success": true, "message": "Item updated successfully"})
}

func UpdateItem2(_ context.Context, c *app.RequestContext) {
	var updatedItem conf.UpdateItem
	c.Bind(&updatedItem)
	if err := c.Bind(&updatedItem); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 返回 JWT 密钥
		return conf.Con.Jwtkey, nil
	})
	if err != nil || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 连接数据库
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conf.DB}))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}

	// 更新数据库中的记录
	toUpdate := conf.Item{
		Name:      updatedItem.Name,
		SimpleDsc: updatedItem.Description,
		Price:     int(updatedItem.Price),
		Dsc:       updatedItem.Dsc,
		OnSale:    updatedItem.Sale,
	}
	// recommendationUpdate := gorseCli.Item{}
	// conf.GorseClient.UpdateItem()
	updateResult := db.Where("id = ?", updatedItem.ID).Select("name", "simple_dsc", "price", "dsc", "on_sale").Updates(&toUpdate)
	if updateResult.Error != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database update error"})
		return
	}
	var updated conf.Item
	if db.Model(&toUpdate).Where("id = ?", updatedItem.ID).Select("on_sale, decision").Scan(&updated).Error == nil {
		// 更新推荐系统中的记录
		conf.SetItemHidden(strconv.Itoa(updatedItem.ID), !(updated.OnSale && updated.Decision == "APPROVE"))
	}

	// 返回成功响应
	c.JSON(http.StatusOK, utils.H{"success": true, "message": "Item updated successfully"})
}
