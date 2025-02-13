package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type Paging struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

func Display(_ context.Context, c *app.RequestContext) {
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}
	defer db.Close() // 确保数据库连接在结束时关闭

	var page Paging
	page.Page = 1
	page.PageSize = 10

	err = c.Bind(&page)
	if err != nil {
		log.Fatal("Bind parameter error")
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Bind parameter error"})
		return
	}

	// 构建基础查询 SQL
	baseQuery := "SELECT id, name, simple_dsc, owner, price, img FROM item WHERE on_sale = 1"
	var query string
	var rows *sql.Rows
	var totalItems int

	query = baseQuery + " LIMIT ? OFFSET ?"

	// 查询总项目数 (用于计算 totalPages)
	countQuery := "SELECT COUNT(*) FROM item WHERE on_sale = 1"
	err = db.QueryRow(countQuery).Scan(&totalItems)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database count query error"})
		return
	}

	rows, err = db.Query(query, page.PageSize, (page.Page-1)*page.PageSize) // 注意参数顺序

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
		var img string

		// 扫描数据
		if err := rows.Scan(&id, &name, &simple_des, &owner, &price, &img); err != nil {
			log.Fatal(err)
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

	// 计算总页数
	totalPages := (totalItems + page.PageSize - 1) / page.PageSize

	// 返回结果，包含 totalPages
	c.JSON(http.StatusOK, utils.H{
		"items":      items,
		"totalPages": totalPages, // 添加 totalPages
	})
}
