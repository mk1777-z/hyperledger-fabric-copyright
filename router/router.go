package router

import (
	"hyperledger-fabric-copyright/middle"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(h *server.Hertz) {
	// API路由前缀组
	api := h.Group("/api")
	{
		// 项目列表API
		api.POST("/items", middle.GetItems)

		// 项目详情API
		api.POST("/information", middle.GetItemInfo)

		// 搜索API
		api.POST("/search", middle.SearchItems)

		// 收藏相关API
		api.GET("/favorites", middle.GetFavorites)
		api.POST("/favorites/add", middle.AddFavorite)
		api.POST("/favorites/remove", middle.RemoveFavorite)

		// 图片代理API
		api.GET("/proxy/image", middle.ProxyImage)

		// 聊天功能路由组
		chatApi := api.Group("/chat")
		{
			chatApi.POST("/send", middle.SendMessageHandler)
			chatApi.GET("/messages/:conversation_id", middle.GetMessagesHandler)
			chatApi.GET("/conversations", middle.GetConversationsHandler)
			chatApi.POST("/messages/read", middle.MarkAsReadHandler)
		}
	}

	// 统计分析相关路由
	h.POST("/statistics/chartData", middle.GetChartData)
	h.POST("/statistics/exportData", middle.ExportData)

	// 添加数据源信息检查接口
	h.GET("/api/statistics/datasources", middle.GetDataSourceInfo)

	// 其他路由注册...
}
