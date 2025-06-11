package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

// AddFavorite 添加收藏
func AddFavorite(_ context.Context, c *app.RequestContext) {
	var req struct {
		ItemID int `json:"item_id"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效的请求参数"})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "缺少授权令牌"})
		return
	}
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌"})
		return
	}
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌声明"})
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库连接错误"})
		return
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM item WHERE id = ?)", req.ItemID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库查询错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, utils.H{"message": "商品不存在"})
		return
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM favorites WHERE username = ? AND item_id = ?)",
		claims.Username, req.ItemID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库查询错误"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, utils.H{"message": "已经收藏过该商品"})
		return
	}

	_, err = db.Exec("INSERT INTO favorites (username, item_id, create_time) VALUES (?, ?, ?)",
		claims.Username, req.ItemID, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "添加收藏失败"})
		return
	}

	c.JSON(http.StatusOK, utils.H{"message": "收藏成功"})
}

// RemoveFavorite 取消收藏
func RemoveFavorite(_ context.Context, c *app.RequestContext) {
	var req struct {
		ItemID int `json:"item_id"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效的请求参数"})
		return
	}
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "缺少授权令牌"})
		return
	}
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌"})
		return
	}
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌声明"})
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库连接错误"})
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM favorites WHERE username = ? AND item_id = ?",
		claims.Username, req.ItemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "取消收藏失败"})
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "查询删除结果失败"})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, utils.H{"message": "未找到该收藏记录"})
		return
	}

	c.JSON(http.StatusOK, utils.H{"message": "取消收藏成功"})
}

// GetFavorites 获取用户的收藏列表，包括item表详情
func GetFavorites(_ context.Context, c *app.RequestContext) {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) == 0 {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "缺少授权令牌"})
		return
	}
	tokenStr := strings.Replace(string(tokenString), "Bearer ", "", -1)
	token, err := jwt.ParseWithClaims(tokenStr, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌"})
		return
	}
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌声明"})
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库连接错误"})
		return
	}
	defer db.Close()

	// 查询收藏 + item 信息
	rows, err := db.Query(`SELECT i.id, i.name, i.simple_dsc, i.owner, i.price, i.img, f.create_time 
		FROM favorites f 
		JOIN item i ON f.item_id = i.id 
		WHERE f.username = ? ORDER BY f.create_time DESC`, claims.Username)
	if err != nil {
		log.Printf("[QUERY ERROR] username=%s err=%v", claims.Username, err)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "查询收藏失败"})
		return
	}
	defer rows.Close()

	// 使用可空类型
	type FavoriteItem struct {
		ID         int            `json:"id"`
		Name       sql.NullString `json:"name"`
		SimpleDsc  sql.NullString `json:"simple_dsc"`
		Owner      sql.NullString `json:"owner"`
		PriceStr   sql.NullString `json:"price"`
		Img        sql.NullString `json:"img"`
		CreateTime sql.NullString `json:"create_time"`
	}

	var favorites []map[string]interface{}
	for rows.Next() {
		var fi FavoriteItem
		if err := rows.Scan(&fi.ID, &fi.Name, &fi.SimpleDsc, &fi.Owner, &fi.PriceStr, &fi.Img, &fi.CreateTime); err != nil {
			log.Printf("[SCAN ERROR] id=%d err=%v", fi.ID, err)
			c.JSON(http.StatusInternalServerError, utils.H{"message": "处理查询结果失败"})
			return
		}
		favorites = append(favorites, map[string]interface{}{
			"id":          fi.ID,
			"name":        nullableString(fi.Name),
			"simple_dsc":  nullableString(fi.SimpleDsc),
			"owner":       nullableString(fi.Owner),
			"price":       parsePrice(fi.PriceStr),
			"img":         nullableString(fi.Img),
			"create_time": nullableString(fi.CreateTime),
		})
	}

	c.JSON(http.StatusOK, utils.H{
		"message": "获取收藏列表成功",
		"data":    favorites,
	})
}

func nullableString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func parsePrice(ns sql.NullString) float64 {
	if ns.Valid {
		if f, err := strconv.ParseFloat(ns.String, 64); err == nil {
			return f
		}
	}
	return 0
}

func nullableFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}
