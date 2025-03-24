package middle

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// 定义一个简单的H类型替代utils.H
type H map[string]interface{}

// 统计分析页面
func StatisticsPage(ctx context.Context, c *app.RequestContext) {
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

	// 从数据库获取数据
	categoryData, err := getCategoryData(req.TimeRange, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting category data"})
		return
	}

	priceData, err := getPriceData(req.TimeRange, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting price data"})
		return
	}

	trendData, err := getTrendData(req.TimeRange, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting trend data"})
		return
	}

	activityData, err := getActivityData(req.TimeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting activity data"})
		return
	}

	c.JSON(http.StatusOK, H{
		"categoryData": categoryData,
		"priceData":    priceData,
		"trendData":    trendData,
		"activityData": activityData,
	})
}

// 获取交易数据
func GetTransactionData(ctx context.Context, c *app.RequestContext) {
	var req struct {
		TimeRange string `json:"timeRange"`
		Category  string `json:"category"`
		Page      int    `json:"page"`
		Limit     int    `json:"limit"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, H{"message": "Invalid request"})
		return
	}

	// 从数据库获取数据
	data, count, err := getTransactionData(req.TimeRange, req.Category, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting transaction data"})
		return
	}

	c.JSON(http.StatusOK, H{
		"code":  0,
		"msg":   "",
		"count": count,
		"data":  data,
	})
}

// 获取用户数据
func GetUserData(ctx context.Context, c *app.RequestContext) {
	var req struct {
		TimeRange string `json:"timeRange"`
		Page      int    `json:"page"`
		Limit     int    `json:"limit"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, H{"message": "Invalid request"})
		return
	}

	// 从数据库获取数据
	data, count, err := getUserData(req.TimeRange, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting user data"})
		return
	}

	c.JSON(http.StatusOK, H{
		"code":  0,
		"msg":   "",
		"count": count,
		"data":  data,
	})
}

// 获取收益数据
func GetRevenueData(ctx context.Context, c *app.RequestContext) {
	var req struct {
		TimeRange string `json:"timeRange"`
		Page      int    `json:"page"`
		Limit     int    `json:"limit"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, H{"message": "Invalid request"})
		return
	}

	// 从数据库获取数据
	data, count, err := getRevenueData(req.TimeRange, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting revenue data"})
		return
	}

	c.JSON(http.StatusOK, H{
		"code":  0,
		"msg":   "",
		"count": count,
		"data":  data,
	})
}

// 获取地理位置数据
func GetLocationData(ctx context.Context, c *app.RequestContext) {
	// 从数据库获取数据
	data, err := getLocationData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"message": "Error getting location data"})
		return
	}

	c.JSON(http.StatusOK, data)
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
			c.JSON(http.StatusInternalServerError, H{"message": "Error exporting Excel"})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=copyright_statistics.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file)
	} else if req.Type == "pdf" {
		// 导出PDF
		file, err := exportPDF(req.TimeRange, req.Category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, H{"message": "Error exporting PDF"})
			return
		}

		c.Header("Content-Disposition", "attachment; filename=copyright_statistics.pdf")
		c.Data(http.StatusOK, "application/pdf", file)
	} else {
		c.JSON(http.StatusBadRequest, H{"message": "Unsupported export type"})
	}
}

// 辅助函数 - 获取分类数据
func getCategoryData(timeRange, category string) ([]map[string]interface{}, error) {
	// 这里实现从数据库获取分类统计数据的逻辑
	// 示例数据
	data := []map[string]interface{}{
		{"name": "音乐", "value": 120},
		{"name": "图片", "value": 85},
		{"name": "视频", "value": 65},
		{"name": "文档", "value": 50},
		{"name": "其他", "value": 30},
	}
	return data, nil
}

// 辅助函数 - 获取价格分布数据
func getPriceData(timeRange, category string) ([]map[string]interface{}, error) {
	// 示例数据
	data := []map[string]interface{}{
		{"range": "0-100", "count": 120},
		{"range": "100-500", "count": 85},
		{"range": "500-1000", "count": 65},
		{"range": "1000-5000", "count": 30},
		{"range": "5000以上", "count": 10},
	}
	return data, nil
}

// 辅助函数 - 获取趋势数据
func getTrendData(timeRange, category string) (map[string]interface{}, error) {
	// 示例数据
	now := time.Now()
	months := []string{}
	counts := []int{}
	amounts := []float64{}

	// 生成过去6个月的数据
	for i := 5; i >= 0; i-- {
		month := now.AddDate(0, -i, 0).Format("2006-01")
		months = append(months, month)
		counts = append(counts, 10+i*5)
		amounts = append(amounts, 1000.0+float64(i*500))
	}

	data := map[string]interface{}{
		"months":  months,
		"counts":  counts,
		"amounts": amounts,
	}
	return data, nil
}

// 辅助函数 - 获取活跃度数据
func getActivityData(timeRange string) (map[string]interface{}, error) {
	// 示例数据
	now := time.Now()
	dates := []string{}
	newUsers := []int{}
	activeUsers := []int{}
	tradingUsers := []int{}

	// 生成过去7天的数据
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("01-02")
		dates = append(dates, date)
		newUsers = append(newUsers, 5+i)
		activeUsers = append(activeUsers, 20+i*2)
		tradingUsers = append(tradingUsers, 10+i)
	}

	data := map[string]interface{}{
		"dates":        dates,
		"newUsers":     newUsers,
		"activeUsers":  activeUsers,
		"tradingUsers": tradingUsers,
	}
	return data, nil
}

// 辅助函数 - 获取交易数据表格
func getTransactionData(timeRange, category string, page, limit int) ([]map[string]interface{}, int, error) {
	// 示例数据
	transactions := []map[string]interface{}{}

	// 生成示例记录
	for i := 1; i <= limit; i++ {
		id := (page-1)*limit + i
		transactions = append(transactions, map[string]interface{}{
			"id":       id,
			"name":     "版权商品-" + string(rune(64+id)),
			"category": []string{"音乐", "图片", "视频", "文档", "其他"}[id%5],
			"seller":   "卖家-" + string(rune(64+id%10)),
			"buyer":    "买家-" + string(rune(64+id%8)),
			"price":    100.0 + float64(id*10),
			"time":     time.Now().AddDate(0, 0, -id).Format("2006-01-02 15:04:05"),
			"location": "北京市朝阳区",
		})
	}

	return transactions, 100, nil // 总记录数暂定为100
}

// 辅助函数 - 获取用户数据表格
func getUserData(timeRange string, page, limit int) ([]map[string]interface{}, int, error) {
	// 示例数据
	users := []map[string]interface{}{}

	// 生成示例记录
	for i := 1; i <= limit; i++ {
		id := (page-1)*limit + i
		users = append(users, map[string]interface{}{
			"username":       "用户-" + string(rune(64+id)),
			"buyCount":       id % 10,
			"sellCount":      id % 8,
			"totalAmount":    1000.0 + float64(id*100),
			"lastActiveTime": time.Now().AddDate(0, 0, -id%30).Format("2006-01-02 15:04:05"),
		})
	}

	return users, 80, nil // 总记录数暂定为80
}

// 辅助函数 - 获取收益数据表格
func getRevenueData(timeRange string, page, limit int) ([]map[string]interface{}, int, error) {
	// 示例数据
	revenues := []map[string]interface{}{}

	// 生成示例记录
	for i := 1; i <= limit; i++ {
		id := (page-1)*limit + i
		month := time.Now().AddDate(0, -id, 0).Format("2006-01")
		revenues = append(revenues, map[string]interface{}{
			"period":           month,
			"totalRevenue":     10000.0 + float64(id*1000),
			"transactionCount": 50 + id*5,
			"avgPrice":         200.0 + float64(id*10),
			"growth":           5.0 + float64(id%10),
		})
	}

	return revenues, 24, nil // 总记录数暂定为24（两年）
}

// 辅助函数 - 获取地理位置数据
func getLocationData() ([]map[string]interface{}, error) {
	// 示例数据 - 中国主要城市坐标
	locations := []map[string]interface{}{
		{"username": "用户A", "lng": 116.407, "lat": 39.904, "count": 10, "lastTransaction": "2023-10-01"}, // 北京
		{"username": "用户B", "lng": 121.473, "lat": 31.230, "count": 15, "lastTransaction": "2023-10-05"}, // 上海
		{"username": "用户C", "lng": 113.280, "lat": 23.125, "count": 8, "lastTransaction": "2023-10-08"},  // 广州
		{"username": "用户D", "lng": 114.085, "lat": 22.547, "count": 12, "lastTransaction": "2023-10-10"}, // 深圳
		{"username": "用户E", "lng": 104.066, "lat": 30.659, "count": 6, "lastTransaction": "2023-10-12"},  // 成都
	}

	return locations, nil
}

// 辅助函数 - 导出Excel
func exportExcel(timeRange, category string) ([]byte, error) {
	// 模拟Excel文件内容
	return []byte("Excel文件示例内容"), nil
}

// 辅助函数 - 导出PDF
func exportPDF(timeRange, category string) ([]byte, error) {
	// 模拟PDF文件内容
	return []byte("PDF文件示例内容"), nil
}
