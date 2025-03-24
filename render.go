package main

import (
	"context"
	"hyperledger-fabric-copyright/middle" // 添加这行导入middle包

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func renderHTML(h *server.Hertz) {
	// 加载HTML模板文件
	h.LoadHTMLGlob("HTML/project/*")

	h.Static("/static", "./")

	// 默认根路径返回一个 JSON 响应
	h.GET("/homepage", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "homepage.html", utils.H{
			"title": "Home",
		})
	})

	// 渲染 signin 页面
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "signin.html", utils.H{
			"title": "Sign In",
		})
	})

	// 渲染 signup 页面
	h.GET("/signup", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "signup.html", utils.H{
			"title": "Sign Up",
		})
	})

	h.GET("/information", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "information.html", utils.H{
			"title": "Information",
		})
	})
	h.GET("/display", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "display.html", utils.H{
			"title": "display",
		})
	})
	h.GET("/upload", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "upload.html", utils.H{
			"title": "upload",
		})
	})

	// 添加对statistics页面的支持
	h.GET("/statistics", middle.StatisticsPage)
}
