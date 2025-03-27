package middle

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// GetItems 是 Display 函数的包装器，用于获取版权项目列表
func GetItems(ctx context.Context, c *app.RequestContext) {
	Display(ctx, c)
}

// GetItemInfo 是 Information 函数的包装器，用于获取版权项目详情
func GetItemInfo(ctx context.Context, c *app.RequestContext) {
	Information(ctx, c)
}

// SearchItems 是 Search 函数的包装器，用于搜索版权项目
func SearchItems(ctx context.Context, c *app.RequestContext) {
	Search(ctx, c)
}
