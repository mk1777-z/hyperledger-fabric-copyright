package middle

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

// Evaluate a transaction by assetID to query ledger state.
func readAssetByID(contract *client.Contract, assetId string) {
	fmt.Printf("\n--> Evaluate Transaction: ReadAsset, function returns asset attributes\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadCreatetrans", assetId)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

func Information(_ context.Context, c *app.RequestContext) {
	var body RequestBody
	if err := c.Bind(&body); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body"})
		return
	}

	itemID := body.Name
	// 从请求中获取商品 ID
	if itemID == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Item name is missing"})
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
	defer db.Close() // 确保数据库连接在结束时关闭

	// 查询数据库，获取该用户的项目列表
	rows, err := db.Query("SELECT id, name, simple_dsc, price, dsc, owner, img, start_time,transID FROM item WHERE name = ?", itemID)
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
		var name, simple_des, owner, dsc, img, start_time string
		var transID *string
		var price float32
		if err := rows.Scan(&id, &name, &simple_des, &price, &dsc, &owner, &img, &start_time, &transID); err != nil {
			c.Status(http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, utils.H{"message": "Error reading row"})
			return
		}
		trace := strings.Split(*transID, " ")
		for i := 0; i < len(trace); i++ {
			readAssetByID(conf.Contract, trace[i])
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": simple_des,
			"price":       price,
			"owner":       owner,
			"dsc":         dsc,
			"img":         img,
			"start_time":  start_time,
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"message": "Items fetched successfully",
		"items":   items,
	})
}
