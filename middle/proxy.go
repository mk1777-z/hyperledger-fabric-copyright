package middle

import (
	"context"
	"io"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// ProxyImage 代理获取图片
func ProxyImage(_ context.Context, c *app.RequestContext) {
	imageID := c.Query("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "图片ID不能为空",
		})
		return
	}

	// 构建华为云OBS的URL
	obsURL := "https://huaweibucket-48f4.obs.cn-east-3.myhuaweicloud.com/" + imageID

	// 发起请求获取图片
	resp, err := http.Get(obsURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "获取图片失败",
		})
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cache-Control", "public, max-age=31536000")

	// 将图片数据写入响应
	io.Copy(c.Response.BodyWriter(), resp.Body)
}
