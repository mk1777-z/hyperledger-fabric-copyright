package main

import (
	"hyperledger-fabric-copyright/conf"
	"hyperledger-fabric-copyright/middle"

	"github.com/cloudwego/hertz/pkg/app/server"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	h := server.Default()

	renderHTML(h)

	conf.Init()

	h.POST("/register", middle.Register)

	h.POST("/login", middle.Login)

	h.POST("/myproject", middle.Myproject)

	h.POST("/display", middle.Display)

	h.POST("/upload", middle.Upload)

	h.POST("/information", middle.Information)

	h.POST("/updateItem", middle.UpdateItem)

	h.POST("/transaction", middle.Transaction)

	h.POST("/account", middle.HandleAccount)

	h.GET("/chat_ws", middle.ChatWebsocket)
	h.NoHijackConnPool = true
	h.Spin()
}
