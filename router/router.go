package router

import (
	"hyperledger-fabric-copyright/middle" // 修改这里，移除github.com/前缀

	"github.com/cloudwego/hertz/pkg/app/server"
)

func RegisterRoutes(h *server.Hertz) {
	// 注释掉这个路由，因为它已经在renderHTML中定义了
	// h.GET("/", func(ctx context.Context, c *app.RequestContext) {
	//     c.HTML(http.StatusOK, "signin.html", nil)
	// })

	// 注册统计分析相关路由
	// h.GET("/statistics", middle.StatisticsPage) // 删除，因为它已经在renderHTML中定义了
	h.POST("/statistics/chartData", middle.GetChartData)
	h.POST("/statistics/transactionData", middle.GetTransactionData)
	h.POST("/statistics/userData", middle.GetUserData)
	h.POST("/statistics/revenueData", middle.GetRevenueData)
	h.POST("/statistics/locationData", middle.GetLocationData)
	h.POST("/statistics/exportData", middle.ExportData)

	// 其他路由注册...
}
