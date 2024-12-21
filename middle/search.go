package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type RequestBody struct {
	Name string `json:"name"`
}

func Search(_ context.Context, c *app.RequestContext) {
	var body RequestBody
	if err := c.Bind(&body); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body"})
		return
	}

	name := body.Name
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}
	defer db.Close() // 确保数据库连接在结束时关闭
	rows, err := db.Query("SELECT id, name, simple_dsc, price, owner, dsc, img, on_sale, start_time FROM item WHERE name LIKE CONCAT('%', ?, '%')", name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database query error"})
		return
	}
	defer rows.Close()
	if name == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Name parameter is required"})
		return
	}

	// 存储查询结果
	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var name, simple_des, owner, dsc, img, start_time string
		var price float32
		var on_sale bool
		if err := rows.Scan(&id, &name, &simple_des, &price, &owner, &dsc, &img, &on_sale, &start_time); err != nil {
			c.Status(http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, utils.H{"message": "Error reading row"})
			return
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": simple_des,
			"price":       price,
		})
	}
	c.JSON(http.StatusOK, utils.H{
		"message": "Items fetched successfully",
		"items":   items,
	})
}
