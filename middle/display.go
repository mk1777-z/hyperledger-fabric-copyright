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
	page.PageSize = 12 // 设置为12个项目每页

	err = c.Bind(&page)
	if err != nil {
		log.Fatal("Bind parameter error")
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Bind parameter error"})
		return
	}

	// 计算分页的偏移量
	offset := (page.Page - 1) * page.PageSize

	// 直接查询decision字段为APPROVE的在售项目
	countQuery := "SELECT COUNT(*) FROM item WHERE on_sale = 1 AND decision = 'APPROVE'"
	var totalItems int
	err = db.QueryRow(countQuery).Scan(&totalItems)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database count query error"})
		return
	}

	// 查询筛选后的项目，并应用分页
	itemsQuery := "SELECT id, name, simple_dsc, owner, price, img FROM item WHERE on_sale = 1 AND decision = 'APPROVE' LIMIT ? OFFSET ?"
	rows, err := db.Query(itemsQuery, page.PageSize, offset)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database query error"})
		return
	}
	defer rows.Close()

	// 存储查询结果
	var approvedItems []map[string]interface{}
	for rows.Next() {
		var id int
		var name, simple_des string
		var price float32
		var owner string
		var img string

		// 扫描数据
		if err := rows.Scan(&id, &name, &simple_des, &owner, &price, &img); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // 跳过此行，继续处理下一行
		}

		if img == "NULL" {
			img = "noimage"
		}

		// 创建项目对象
		item := map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": simple_des,
			"price":       price,
			"img":         img,
			"owner":       owner,
		}

		approvedItems = append(approvedItems, item)
		// 移除了添加到结果中的日志
	}

	// 计算总页数
	totalPages := (totalItems + page.PageSize - 1) / page.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	log.Printf("返回第 %d 页项目，每页 %d 个，总项目数: %d, 总页数: %d",
		page.Page, page.PageSize, totalItems, totalPages)

	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"items":      approvedItems,
		"totalPages": totalPages,
		"totalItems": totalItems,
	})
}
