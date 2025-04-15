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
	// 修改模板加载方式，只加载HTML文件
	h.LoadHTMLGlob("HTML/project/*.html")

	// 配置静态文件目录，使图片和其他资源文件可以正常访问
	h.Static("/HTML", "./HTML")

	// 默认根路径返回一个 JSON 响应
	h.GET("/homepage", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "homepage.html", utils.H{
			"title": "Home",
		})
	})

	h.GET("/search", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "search.html", utils.H{
			"title": "Search",
		})
	})

	// 将根路径改为指向home.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "home.html", utils.H{
			"title": "区块链版权管理系统",
		})
	})

	// 添加新路由指向登录页面
	h.GET("/login", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "signin.html", utils.H{
			"title": "Sign In",
		})
	})

	// 渲染 signup 页面
	h.GET("/register", func(ctx context.Context, c *app.RequestContext) {
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

	//炫酷的主页
	h.GET("/home", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "home.html", utils.H{
			"title": "区块链版权管理系统",
		})
	})
}
