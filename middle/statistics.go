package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// 定义一个简单的H类型替代utils.H
type H map[string]interface{}

// 统计分析页面
func StatisticsPage(ctx context.Context, c *app.RequestContext) {
	// 在模板加载时，Hertz引擎会省略project前缀，直接使用文件名
	c.HTML(http.StatusOK, "statistics.html", nil)
}

// 获取图表数据
func GetChartData(ctx context.Context, c *app.RequestContext) {
	var req struct {
		TimeRange string `json:"timeRange"`
		Category  string `json:"category"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, H{"message": "Invalid request"})
		return
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Database connection error: " + err.Error()})
		return
	}
	defer db.Close()

	// 从数据库获取数据
	categoryData, err := getCategoryData(db, req.TimeRange, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting category data: " + err.Error()})
		return
	}

	priceData, err := getPriceData(db, req.TimeRange, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting price data: " + err.Error()})
		return
	}

	trendData, err := getTrendData(db, req.TimeRange, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting trend data: " + err.Error()})
		return
	}

	activityData, err := getActivityData(db, req.TimeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting activity data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, H{
		"categoryData": categoryData,
		"priceData":    priceData,
		"trendData":    trendData,
		"activityData": activityData,
	})
}

// 导出数据
func ExportData(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Type      string `json:"type"`
		TimeRange string `json:"timeRange"`
		Category  string `json:"category"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, H{"message": "Invalid request"})
		return
	}

	// 根据类型导出不同格式
	if req.Type == "excel" {
		// 导出Excel
		file, err := exportExcel(req.TimeRange, req.Category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, H{"message": "Error exporting Excel: " + err.Error()})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=copyright_statistics.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file)
	} else if req.Type == "pdf" {
		// 导出PDF
		file, err := exportPDF(req.TimeRange, req.Category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, H{"message": "Error exporting PDF: " + err.Error()})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=copyright_statistics.pdf")
		c.Data(http.StatusOK, "application/pdf", file)
	} else {
		c.JSON(http.StatusBadRequest, H{"message": "Unsupported export type"})
	}
}

// 解析时间范围，返回开始和结束时间
func parseTimeRange(timeRange string) (time.Time, time.Time, error) {
	startTime := time.Now().AddDate(0, -1, 0) // 默认过去30天
	endTime := time.Now()

	if timeRange != "" {
		// 假设timeRange格式为"2023-01-01 - 2023-01-31"
		parts := strings.Split(timeRange, " - ")
		if len(parts) == 2 {
			var err error
			startTime, err = time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
			if err != nil {
				return startTime, endTime, err
			}
			endTime, err = time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
			if err != nil {
				return startTime, endTime, err
			}
			// 将结束时间设置为当天的最后一刻
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
		}
	}

	return startTime, endTime, nil
}

// 从数据库获取版权分类统计数据
func getCategoryData(db *sql.DB, timeRange, category string) ([]map[string]interface{}, error) {
	startTime, endTime, err := parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	// 修改SQL查询 - 使用正确的start_time字段而不是created_at
	query := `
		SELECT 
			category as name, 
			COUNT(*) as value
		FROM item
		WHERE start_time BETWEEN ? AND ?
	`

	params := []interface{}{startTime, endTime}

	if category != "" {
		query += " AND category = ?"
		params = append(params, category)
	}

	query += " GROUP BY category ORDER BY value DESC"

	// 执行查询
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("查询分类数据失败: %v", err)
	}
	defer rows.Close()

	// 处理结果
	var result []map[string]interface{}
	for rows.Next() {
		var name string
		var value int

		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}

	// 如果没有数据，返回基本分类数据
	if len(result) == 0 {
		// 获取所有不同的分类
		query = "SELECT DISTINCT category FROM item"
		rows, err = db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("获取分类列表失败: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			result = append(result, map[string]interface{}{
				"name":  name,
				"value": 0,
			})
		}
	}

	// 如果结果仍然为空，返回错误
	if len(result) == 0 {
		return nil, fmt.Errorf("未找到分类数据")
	}

	return result, nil
}

// 从数据库获取价格分布数据
func getPriceData(db *sql.DB, timeRange, category string) ([]map[string]interface{}, error) {
	// 定义价格区间
	priceRanges := []struct {
		Min       int
		Max       int
		RangeName string
	}{
		{0, 100, "0-100"},
		{100, 500, "100-500"},
		{500, 1000, "500-1000"},
		{1000, 5000, "1000-5000"},
		{5000, 999999, "5000以上"},
	}

	// 构建结果
	var result []map[string]interface{}
	var hasData bool

	for _, r := range priceRanges {
		// 修改查询，直接使用category字段
		query := `
			SELECT COUNT(*) 
			FROM item 
			WHERE price >= ? AND price < ?
		`
		params := []interface{}{r.Min, r.Max}

		if category != "" {
			query += ` AND category = ?`
			params = append(params, category)
		}

		var count int
		err := db.QueryRow(query, params...).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("查询价格区间 %s 失败: %v", r.RangeName, err)
		}

		if count > 0 {
			hasData = true
		}

		result = append(result, map[string]interface{}{
			"range": r.RangeName,
			"count": count,
		})
	}

	if !hasData {
		return nil, fmt.Errorf("未找到价格分布数据")
	}

	return result, nil
}

// 从区块链读取交易数据 - 重命名以避免与information.go中的函数冲突
func readTransactionFromBlockchainForStats() ([]map[string]interface{}, error) {
	log.Println("从区块链读取交易数据...")

	// 连接数据库获取所有项目的 transID
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 查询所有有交易ID的项目
	rows, err := db.Query("SELECT name, transID FROM item WHERE transID IS NOT NULL AND transID != ''")
	if err != nil {
		return nil, fmt.Errorf("查询项目交易ID失败: %v", err)
	}
	defer rows.Close()

	var allTransactions []map[string]interface{}

	// 遍历每个项目
	for rows.Next() {
		var name string
		var transID string
		if err := rows.Scan(&name, &transID); err != nil {
			log.Printf("扫描行数据失败: %v", err)
			continue
		}

		// 跳过空的 transID
		if transID == "" {
			continue
		}

		// 解析 transID 字符串，获取每个交易的 assetID
		trace := strings.Split(transID, " ")
		for _, assetID := range trace {
			// 从区块链读取交易详情 - 改为使用information中已定义的函数
			transactionDetails, err := readAssetByID(conf.BasicContract, assetID)
			if err != nil {
				log.Printf("获取交易 %s 详情失败: %v", assetID, err)
				continue
			}

			// 添加项目名称到交易详情
			transactionDetails["ItemName"] = name

			// 将交易添加到结果中
			allTransactions = append(allTransactions, transactionDetails)
		}
	}

	log.Printf("成功从区块链读取 %d 条交易记录", len(allTransactions))
	return allTransactions, nil
}

// 从数据库获取交易趋势数据
func getTrendData(db *sql.DB, timeRange, category string) (map[string]interface{}, error) {
	startTime, endTime, err := parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	// 计算月份范围（最多6个月）
	sixMonthsAgo := time.Now().AddDate(0, -5, 0).Format("2006-01")
	if startTime.Format("2006-01") < sixMonthsAgo {
		startTime, _ = time.Parse("2006-01", sixMonthsAgo)
	}

	// 从区块链获取所有交易数据
	allTransactions, err := readTransactionFromBlockchainForStats()
	if err != nil {
		log.Printf("获取区块链交易数据失败: %v", err)
		return nil, fmt.Errorf("获取区块链交易数据失败: %v", err)
	}

	// 确保每个月都有数据
	monthMap := make(map[string]bool)
	for current := startTime; !current.After(endTime); current = current.AddDate(0, 1, 0) {
		monthMap[current.Format("2006-01")] = false
	}

	// 按月份统计交易数据
	monthData := make(map[string]struct {
		count  int
		amount float64
	})

	// 如果需要按类别筛选，先获取该类别的所有项目名称
	var categoryItems []string
	if category != "" {
		categoryItemsRows, err := db.Query("SELECT name FROM item WHERE category = ?", category)
		if err == nil {
			defer categoryItemsRows.Close()
			for categoryItemsRows.Next() {
				var name string
				if err := categoryItemsRows.Scan(&name); err == nil {
					categoryItems = append(categoryItems, name)
				}
			}
		}
	}

	// 处理每个交易
	for _, transaction := range allTransactions {
		// 获取交易时间
		transTime, ok := transaction["Transtime"].(string)
		if !ok {
			continue
		}

		// 解析交易时间
		txTime, err := time.Parse("2006-01-02 15:04:05", transTime)
		if err != nil {
			continue
		}

		// 检查时间范围
		if txTime.Before(startTime) || txTime.After(endTime) {
			continue
		}

		// 如果需要按类别筛选
		if category != "" {
			itemName, ok := transaction["ItemName"].(string)
			if !ok || !contains(categoryItems, itemName) {
				continue
			}
		}

		// 获取交易价格
		var price float64
		switch p := transaction["Price"].(type) {
		case float64:
			price = p
		case string:
			fmt.Sscanf(p, "%f", &price)
		default:
			// 跳过无效价格的交易
			continue
		}

		// 按月份统计
		monthKey := txTime.Format("2006-01")
		monthData[monthKey] = struct {
			count  int
			amount float64
		}{
			count:  monthData[monthKey].count + 1,
			amount: monthData[monthKey].amount + price,
		}
		monthMap[monthKey] = true
	}

	// 整理结果
	var months []string
	var counts []int
	var amounts []float64

	// 将数据转换为数组
	for month := range monthMap {
		months = append(months, month)
	}

	// 按月份排序
	sort.Strings(months)

	// 填充数据
	for _, month := range months {
		data, exists := monthData[month]
		if exists {
			counts = append(counts, data.count)
			amounts = append(amounts, data.amount)
		} else {
			counts = append(counts, 0)
			amounts = append(amounts, 0)
		}
	}

	return map[string]interface{}{
		"months":  months,
		"counts":  counts,
		"amounts": amounts,
	}, nil
}

// 辅助函数：检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 从数据库获取用户活跃度数据
func getActivityData(db *sql.DB, timeRange string) (map[string]interface{}, error) {
	startTime, endTime, err := parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	// 限制为过去7天的数据
	sevenDaysAgo := time.Now().AddDate(0, 0, -6)
	if startTime.Before(sevenDaysAgo) {
		startTime = sevenDaysAgo
	}

	// 构建日期范围
	var dates []string
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("01-02"))
	}

	// 简化处理，直接使用区块链交易数据
	var newUsers = make([]int, len(dates))
	var activeUsers = make([]int, len(dates))
	var tradingUsers = make([]int, len(dates))

	// 从区块链获取交易用户数据
	tradingUsersMap := make(map[string]map[string]bool) // 日期 -> 用户名 -> true

	// 初始化日期映射
	for i, date := range dates {
		tradingUsersMap[date] = make(map[string]bool)
		newUsers[i] = 0    // 初始化为0
		activeUsers[i] = 0 // 初始化为0
	}

	// 从区块链获取交易数据
	allTransactions, err := readTransactionFromBlockchainForStats()
	if err == nil {
		// 按日期统计交易用户
		for _, tx := range allTransactions {
			transTime, ok := tx["Transtime"].(string)
			if !ok {
				continue
			}

			// 解析交易时间
			txTime, err := time.Parse("2006-01-02 15:04:05", transTime)
			if err != nil {
				continue
			}

			// 检查时间是否在范围内
			if txTime.Before(startTime) || txTime.After(endTime) {
				continue
			}

			// 获取日期格式
			dateKey := txTime.Format("01-02")

			// 记录买家和卖家作为活跃用户
			if buyer, ok := tx["Purchaser"].(string); ok {
				tradingUsersMap[dateKey][buyer] = true
			}

			if seller, ok := tx["Seller"].(string); ok {
				tradingUsersMap[dateKey][seller] = true
			}
		}
	}

	// 计算每个日期的交易用户数
	for i, date := range dates {
		tradingUsers[i] = len(tradingUsersMap[date])
		// 注意: user表中没有registration_time字段，所以不能统计新用户数
		// 注意: 没有用户活动日志表，所以使用交易用户数作为活跃用户数
		activeUsers[i] = len(tradingUsersMap[date])
	}

	return map[string]interface{}{
		"dates":        dates,
		"newUsers":     newUsers,
		"activeUsers":  activeUsers,
		"tradingUsers": tradingUsers,
	}, nil
}

// 导出Excel文件
func exportExcel(timeRange, category string) ([]byte, error) {
	// 导出Excel的实现
	// 这是一个简化的实现，返回一个空的Excel文件
	excelContent := []byte("This is a placeholder for Excel content")
	return excelContent, nil
}

// 导出PDF文件
func exportPDF(timeRange, category string) ([]byte, error) {
	// 导出PDF的实现
	// 这是一个简化的实现，返回一个空的PDF文件
	pdfContent := []byte("This is a placeholder for PDF content")
	return pdfContent, nil
}

// GetDataSourceInfo 获取数据来源信息
func GetDataSourceInfo(ctx context.Context, c *app.RequestContext) {
	// 连接数据库以测试连接是否可用
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	dbAvailable := err == nil
	if dbAvailable {
		defer db.Close()
		// 测试数据库连接是否正常
		dbAvailable = db.Ping() == nil
	}

	// 测试区块链连接
	blockchainAvailable := false
	_, err = readTransactionFromBlockchainForStats()
	if err == nil {
		blockchainAvailable = true
	}

	// 准备数据源信息
	dataSources := make(map[string]string)
	if dbAvailable {
		dataSources["categoryData"] = "数据库"
		dataSources["priceData"] = "数据库"
		dataSources["activityData"] = "数据库"
	} else {
		dataSources["categoryData"] = "模拟数据"
		dataSources["priceData"] = "模拟数据"
		dataSources["activityData"] = "模拟数据"
	}

	if blockchainAvailable {
		dataSources["trendData"] = "区块链"
	} else {
		dataSources["trendData"] = "模拟数据"
	}

	c.JSON(http.StatusOK, H{
		"dataStatus": H{
			"db":         dbAvailable,
			"blockchain": blockchainAvailable,
		},
		"dataSources":  dataSources,
		"mockDataUsed": !dbAvailable || !blockchainAvailable,
	})
}
