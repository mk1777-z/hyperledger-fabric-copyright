package main

import (
	"hyperledger-fabric-copyright/conf"
	"hyperledger-fabric-copyright/middle"
	"hyperledger-fabric-copyright/router"
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 服务器启动前先初始化配置
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

	h.POST("/register", middle.Register)

	h.POST("/login", middle.Login)

	h.POST("/myproject", middle.Myproject)

	h.POST("/display", middle.Display)

	h.POST("/upload", middle.Upload)

	h.POST("/information", middle.Information)

	h.POST("/updateItem", middle.UpdateItem)

	h.POST("/transaction", middle.Transaction)

	h.POST("/account", middle.HandleAccount)

	h.POST("/search", middle.Search)

	h.GET("/chat_ws", middle.ChatWebsocket)

	h.NoHijackConnPool = true

	if err := h.Run(); err != nil {
		log.Fatal(err)
	}
}
