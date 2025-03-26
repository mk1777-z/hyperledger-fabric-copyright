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
	_ "github.com/go-sql-driver/mysql" // 确保导入了 MySQL 驱动，如果使用其他数据库请替换
)

// SearchRequest 定义搜索请求的结构体
type SearchRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
}

// SearchResponse 定义搜索响应的结构体
type SearchResponse struct {
	Items      []conf.CopyrightItem `json:"items"`
	TotalPages int                  `json:"totalPages"`
}

func Search(_ context.Context, c *app.RequestContext) {
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}
	defer db.Close() // 确保数据库连接在结束时关闭

	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind request error: %v", err)
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request parameters"})
		return
	}

	page := req.Page
	pageSize := req.PageSize
	searchQuery := req.Search

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15 // 设置默认pageSize
	}

	offset := (page - 1) * pageSize
	searchTerm := "%" + searchQuery + "%"

	// 构建查询数据的 SQL 语句
	query := "SELECT id, name, simple_dsc, owner, price, img FROM item WHERE name LIKE ? OR simple_dsc LIKE ? LIMIT ? OFFSET ?"
	rows, err := db.Query(query, searchTerm, searchTerm, pageSize, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database query failed"})
		return
	}
	defer rows.Close()

	var items []conf.CopyrightItem
	for rows.Next() {
		var item conf.CopyrightItem
		if err := rows.Scan(&item.ID, &item.Name, &item.SimpleDsc, &item.Owner, &item.Price, &item.Img); err != nil {
			log.Printf("Error scanning row: %v", err)
			c.Status(http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, utils.H{"message": "Error processing query results"})
			return
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Error during query results processing"})
		return
	}

	// 构建查询总数的 SQL 语句
	countQuery := "SELECT COUNT(*) FROM item WHERE name LIKE ? OR simple_dsc LIKE ?"
	var totalCount int
	err = db.QueryRow(countQuery, searchTerm, searchTerm).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting total items: %v", err)
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Failed to count total items"})
		return
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	// 返回包含 items 和 totalPages 的 JSON 响应
	c.JSON(http.StatusOK, SearchResponse{
		Items:      items,
		TotalPages: totalPages,
	})
}
