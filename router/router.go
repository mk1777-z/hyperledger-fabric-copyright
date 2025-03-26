package router

import (
	"hyperledger-fabric-copyright/middle"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(h *server.Hertz) {
	// 统计分析相关路由
	h.POST("/statistics/chartData", middle.GetChartData)
	h.POST("/statistics/exportData", middle.ExportData)

	// 添加数据源信息检查接口
	h.GET("/api/statistics/datasources", middle.GetDataSourceInfo)

	// 其他路由注册...
}
