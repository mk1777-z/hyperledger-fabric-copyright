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

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Paging struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type DisplayResponseItem struct {
	Id        int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string `json:"name" gorm:"column:name;type:varchar(50);not null"`
	Owner     string `json:"owner" gorm:"column:owner;type:varchar(255);not null"`
	SimpleDsc string `json:"description" gorm:"column:simple_dsc;type:varchar(30);default:null"`
	Price     int    `json:"price" gorm:"column:price;type:int;not null"`
	Img       string `json:"img" gorm:"column:img;type:longblob;default:null"`
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

func Display2(_ context.Context, c *app.RequestContext) {
	// 通过tokan获取用户名
	tokenBytes := c.GetHeader("Authorization")
	if len(tokenBytes) == 0 {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}
	tokenString := string(tokenBytes)
	if !strings.HasPrefix(tokenString, "Bearer ") {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效的授权头格式"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(tokenString, "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 返回 JWT 密钥
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		// 如果 token 无效，返回 401 未授权错误
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
	username := claims.Username

	var pageData Paging
	pageData.Page = 1
	pageData.PageSize = 12 // 设置为12个项目每页
	// 绑定请求参数
	err = c.Bind(&pageData)
	// 计算分页的偏移量
	offset := (pageData.Page - 1) * pageData.PageSize

	if err != nil {
		log.Fatal("Bind parameter error")
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Bind parameter error"})
		return
	}

	// 获取推荐内容
	gorseClient := conf.GorseClient
	recommandedItems, err := gorseClient.GetRecommendOffSet(context.Background(), username, "", pageData.PageSize, offset)
	recommandedItemsId := make([]int, len(recommandedItems))
	for i := range recommandedItems {
		recommandedItemsId[i], _ = strconv.Atoi(recommandedItems[i])
	}

	// 连接数据库(实际是复用连接)
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conf.DB}))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}
	// 查询总数和推荐的items
	var totalItems int64
	db.Model(&conf.Item{}).Where("on_sale = ? AND decision = ?", 1, "APPROVE").Count(&totalItems)
	itemSlice := []DisplayResponseItem{}
	db.Table("item").Where("id in ?", recommandedItemsId).Select("id", "name", "simple_dsc", "owner", "price", "img").Find(&itemSlice)

	// 计算总页数
	totalPages := (totalItems + int64(pageData.PageSize) - 1) / int64(pageData.PageSize)
	if totalPages == 0 {
		totalPages = 1
	}

	log.Printf("返回第 %d 页项目，每页 %d 个，总项目数: %d, 总页数: %d",
		pageData.Page, pageData.PageSize, totalItems, totalPages)

	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"items":      itemSlice,
		"totalPages": totalPages,
		"totalItems": totalItems,
	})
}
