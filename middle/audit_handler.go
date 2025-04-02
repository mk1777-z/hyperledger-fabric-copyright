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
	"github.com/dgrijalva/jwt-go"
)

// 简化版权项目结构 - 移除了不必要的字段
type AuditItemSimple struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	SimpleDsc string  `json:"simple_dsc"`
	Owner     string  `json:"owner"`
	Price     float64 `json:"price"`
	TransID   string  `json:"transID"`
	StartTime string  `json:"start_time,omitempty"`
	FirstTID  string  `json:"firstTransID"` // 直接存储第一个交易ID，避免前端再处理
}

// 通过审核状态分组的项目集合
type CategorizedItems struct {
	PendingItems  []AuditItemSimple `json:"pendingItems"`
	ApprovedItems []AuditItemSimple `json:"approvedItems"`
	RejectedItems []AuditItemSimple `json:"rejectedItems"`
}

// GetCategorizedItems 获取已分类的版权项目（待审核、已通过、已拒绝）- 优化为不返回图片
func GetCategorizedItems(_ context.Context, c *app.RequestContext) {
	// 获取Token
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
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}

	// 检查用户名是否为监管者
	if claims.Username != "监管者" {
		c.Status(http.StatusForbidden)
		c.JSON(http.StatusForbidden, utils.H{"message": "Only regulators are allowed to access this resource"})
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

	// 只查询必要的列，不查询img列来减轻数据库负担
	rows, err := db.Query("SELECT id, name, simple_dsc, owner, price, transID, start_time FROM item")
	if err != nil {
		log.Printf("查询项目失败: %v", err)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "查询项目失败"})
		return
	}
	defer rows.Close()

	// 存储查询结果
	var allItems []AuditItemSimple
	for rows.Next() {
		var item AuditItemSimple
		var transID, startTime sql.NullString

		if err := rows.Scan(&item.ID, &item.Name, &item.SimpleDsc, &item.Owner, &item.Price, &transID, &startTime); err != nil {
			log.Printf("扫描项目数据失败: %v", err)
			continue
		}

		if transID.Valid {
			item.TransID = transID.String

			// 预处理第一个transID，提前计算好
			transIDs := strings.Split(transID.String, " ")
			if len(transIDs) > 0 && transIDs[0] != "" {
				item.FirstTID = transIDs[0]
			}
		}

		if startTime.Valid {
			item.StartTime = startTime.String
		}

		allItems = append(allItems, item)
	}

	// 按审核状态分类项目
	categorized := CategorizedItems{
		PendingItems:  []AuditItemSimple{},
		ApprovedItems: []AuditItemSimple{},
		RejectedItems: []AuditItemSimple{},
	}

	// 处理每个项目，检查其审核状态
	for _, item := range allItems {
		// 如果没有有效的交易ID，跳过
		if item.FirstTID == "" {
			continue
		}

		// 验证并解析交易ID
		validTradeID, err := validateAndParseTradeID(item.FirstTID)
		if err != nil {
			log.Printf("无效的交易ID %s: %v", item.FirstTID, err)
			continue
		}

		// 检查RegulatorContract是否已正确初始化
		if conf.RegulatorContract == nil {
			log.Printf("RegulatorContract未初始化，无法检查审核状态")
			categorized.PendingItems = append(categorized.PendingItems, item)
			continue
		}

		// 检查交易是否存在
		tradeExists := checkTradeExists(conf.RegulatorContract, validTradeID)
		if !tradeExists {
			// 交易不存在，认为是待审核状态
			categorized.PendingItems = append(categorized.PendingItems, item)
			continue
		}

		// 获取审核历史
		records, err := getAuditHistory(conf.RegulatorContract, validTradeID)
		if err != nil {
			log.Printf("获取审核历史失败 %s: %v", validTradeID, err)
			// 出错时默认为待审核
			categorized.PendingItems = append(categorized.PendingItems, item)
			continue
		}

		// 如果没有审核记录，认为是待审核状态
		if len(records) == 0 {
			categorized.PendingItems = append(categorized.PendingItems, item)
			continue
		}

		// 获取最新的审核记录
		latestRecord := records[len(records)-1]

		// 根据最新记录的决定分类
		switch latestRecord.Decision {
		case "APPROVE":
			categorized.ApprovedItems = append(categorized.ApprovedItems, item)
		case "REJECT":
			categorized.RejectedItems = append(categorized.RejectedItems, item)
		default:
			// 默认为待审核
			categorized.PendingItems = append(categorized.PendingItems, item)
		}
	}

	// 返回分类后的项目
	c.JSON(http.StatusOK, categorized)
}
