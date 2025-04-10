package main

import (
	"context"
	"hyperledger-fabric-copyright/conf"
	"hyperledger-fabric-copyright/middle"
	"hyperledger-fabric-copyright/router"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	_ "github.com/go-sql-driver/mysql"
)

func setupRouter(h *server.Hertz) {
	// 统计分析API - 使用全局路由而非/api前缀
	h.POST("/statistics/summary", middle.GetStatisticsSummaryAPI)
	h.POST("/statistics/copyright-trend", middle.GetCopyrightTrendAPI)
	h.POST("/statistics/copyright-types", middle.GetCopyrightTypesAPI)
	h.POST("/statistics/transaction-amount", middle.GetTransactionAmountAPI)
	h.POST("/statistics/user-activity", middle.GetUserActivityAPI)
	h.POST("/statistics/detail-data", middle.GetDetailTableDataAPI)

	// Excel和PDF导出API
	h.GET("/statistics/export-excel", middle.ExportExcelAPI)
	h.GET("/statistics/export-pdf", middle.ExportPDFAPI)

	// 余额相关API
	h.POST("/balance", middle.HandleAccount)
	h.POST("/getbalance", middle.GetBalance)
	h.POST("/account/balance", middle.GetBalance)

	h.POST("/register", middle.Register)

	h.POST("/login", middle.Login)
	h.POST("/api/login", middle.RegulatorLogin)

	h.POST("/myproject", middle.Myproject)

	h.POST("/display", middle.Display)

	h.POST("/upload", middle.Upload)

	h.POST("/information", middle.Information)

	h.POST("/updateItem", middle.UpdateItem)

	h.POST("/transaction", middle.Transaction)

	h.POST("/search", middle.Search)

	// 审核相关路由
	h.POST("/api/audit", middle.AuditTrade)               // 提交审核决定
	h.GET("/api/audit/history", middle.GetAuditHistory)   // 获取审核历史
	h.GET("/api/audit/info", middle.GetTradeInfoForAudit) // 新增：获取交易信息用于审核

	h.GET("/api/audit/categorized-items", middle.GetCategorizedItems) // 添加新的分类API路由
	// 添加GET请求处理,正确映射audit.html
	h.GET("/audit", func(c context.Context, ctx *app.RequestContext) {
		ctx.HTML(http.StatusOK, "audit.html", nil)
	})

	h.GET("/chat_ws", middle.ChatWebsocket)

	// 交易监控相关路由
	h.GET("/transaction-monitor", func(c context.Context, ctx *app.RequestContext) {
		ctx.HTML(http.StatusOK, "transaction-monitor.html", nil)
	})
	h.GET("/api/transactions", middle.GetAllTransactions)

	// 智能合约说明页面路由
	h.GET("/smart-contracts", func(c context.Context, ctx *app.RequestContext) {
		ctx.HTML(http.StatusOK, "smart-contracts.html", nil)
	})
}

func main() {
	log.Println("正在初始化系统配置...")
	conf.Init()
	log.Println("系统配置初始化完成")

	h := server.Default()

	// 修改静态文件服务根目录，避免路径重复
	h.Static("/static", "/home/hyperledger-fabric-copyright")

	// 只加载project目录下的模板，特别是statistics.html
	h.LoadHTMLGlob("/home/hyperledger-fabric-copyright/HTML/project/*.html")

	// renderHTML函数中不应再注册静态文件
	renderHTML(h)

	// 确保调用路由注册函数
	router.RegisterRoutes(h)

	// 使用setupRouter函数注册API路由
	setupRouter(h)

	h.NoHijackConnPool = true

	if err := h.Run(); err != nil {
		log.Fatal(err)
	}
}
