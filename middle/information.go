package middle

import (
	//"bytes"
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

// Evaluate a transaction by assetID to query ledger state.
func readAssetByID(contract *client.Contract, assetId string) (map[string]interface{}, error) {
	fmt.Printf("\n--> Evaluate Transaction: ReadAsset, function returns asset attributes\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadCreatetrans", assetId)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	// Deserialize the JSON result into a map
	var transactionDetails map[string]interface{}
	err = json.Unmarshal(evaluateResult, &transactionDetails)
	if err != nil {
		panic(fmt.Errorf("failed to parse transaction details: %w", err))
	}

	return transactionDetails, nil
}

// 如果有向前端传递地图相关数据的代码，需要移除或修改
func Information(ctx context.Context, c *app.RequestContext) {
	type RequestBody struct {
		Name string
	}
	var body RequestBody
	if err := c.Bind(&body); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body"})
		return
	}

	itemName := body.Name
	// 从请求中获取商品 ID
	if itemName == "" {
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
	rows, err := db.Query("SELECT id, name, simple_dsc, price, dsc, owner, img, start_time,transID FROM item WHERE name = ?", itemName)
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

		// 初始化交易链上的详细信息
		var transactions []map[string]interface{}

		// 检查 transID 是否为空
		if transID != nil && *transID != "" {
			trace := strings.Split(*transID, " ")
			for _, assetID := range trace {
				transactionDetails, err := readAssetByID(conf.BasicContract, assetID)
				if err != nil {
					fmt.Printf("Failed to fetch transaction details for assetID %s: %v\n", assetID, err)
					continue // 跳过错误的交易记录
				}
				transactions = append(transactions, map[string]interface{}{
					"ID":        transactionDetails["ID"],
					"Name":      transactionDetails["Name"],
					"Price":     fmt.Sprintf("%.2f", transactionDetails["Price"]), // 确保价格是字符串或数值
					"Purchaser": transactionDetails["Purchaser"],
					"Seller":    transactionDetails["Seller"],
					"Transtime": transactionDetails["Transtime"],
				})
			}
		}

		// 将当前结果添加到 items 中
		items = append(items, map[string]interface{}{
			"id":           id,
			"name":         name,
			"description":  simple_des,
			"price":        price,
			"owner":        owner,
			"dsc":          dsc,
			"img":          img,
			"start_time":   start_time,
			"transactions": transactions, // 如果 transID 为空，则 transactions 为空切片
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"message": "Items fetched successfully",
		"items":   items,
	})
}
