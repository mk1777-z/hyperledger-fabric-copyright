package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strings"

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

	// 查询所有在售的项目
	baseQuery := "SELECT id, name, simple_dsc, owner, price, img, transID FROM item WHERE on_sale = 1"
	rows, err := db.Query(baseQuery)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database query error"})
		return
	}
	defer rows.Close()

	// 存储所有通过审核的项目
	var approvedItems []map[string]interface{}
	for rows.Next() {
		var id int
		var name, simple_des string
		var price float32
		var owner string
		var img string
		var transID sql.NullString

		// 扫描数据
		if err := rows.Scan(&id, &name, &simple_des, &owner, &price, &img, &transID); err != nil {
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

		// 检查该项目是否通过了审核
		if transID.Valid && transID.String != "" {
			// 提取以空格分隔的第一个transID
			parts := strings.Split(transID.String, " ")
			if len(parts) > 0 {
				firstTransID := parts[0]

				// 检查这个transID的审核状态
				if isApproved(firstTransID) {
					// 如果审核状态为APPROVE，则添加到结果中
					approvedItems = append(approvedItems, item)
					log.Printf("项目 %s (ID: %d) 通过审核，添加到结果中", name, id)
				} else {
					log.Printf("项目 %s (ID: %d) 未通过审核，不添加到结果中", name, id)
				}
			}
		} else {
			log.Printf("项目 %s (ID: %d) 没有transID或transID为空", name, id)
		}
	}

	// 计算通过审核的项目总数
	approvedItemsCount := len(approvedItems)
	log.Printf("通过审核的项目总数: %d", approvedItemsCount)

	totalPages := (approvedItemsCount + page.PageSize - 1) / page.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// 计算当前页的起始和结束索引
	startIndex := (page.Page - 1) * page.PageSize
	endIndex := startIndex + page.PageSize
	if endIndex > approvedItemsCount {
		endIndex = approvedItemsCount
	}

	// 获取当前页的项目
	var currentPageItems []map[string]interface{}
	if startIndex < approvedItemsCount {
		currentPageItems = approvedItems[startIndex:endIndex]
	} else {
		currentPageItems = []map[string]interface{}{} // 空数组
	}

	log.Printf("返回第 %d 页项目，每页 %d 个，当前页项目数: %d", page.Page, page.PageSize, len(currentPageItems))

	// 返回结果，只包含当前页的通过审核的项目
	c.JSON(http.StatusOK, utils.H{
		"items":      currentPageItems,
		"totalPages": totalPages,
		"totalItems": approvedItemsCount,
	})
}

// 检查交易是否通过审核
func isApproved(tradeID string) bool {
	if conf.RegulatorContract == nil {
		log.Printf("RegulatorContract 未初始化，无法检查审核状态")
		return false
	}

	// 首先检查交易是否存在
	if exists := checkTradeExists(conf.RegulatorContract, tradeID); !exists {
		log.Printf("交易 %s 不存在，未通过审核", tradeID)
		return false
	}

	// 获取审核历史
	records, err := getAuditHistory(conf.RegulatorContract, tradeID)
	if err != nil {
		log.Printf("获取审核历史失败: %v", err)
		return false
	}

	// 如果没有审核记录，则未通过审核
	if len(records) == 0 {
		log.Printf("交易 %s 没有审核记录", tradeID)
		return false
	}

	// 检查最新的审核记录是否批准
	latestRecord := records[len(records)-1] // 获取最新的审核记录
	return latestRecord.Decision == "APPROVE"
}
