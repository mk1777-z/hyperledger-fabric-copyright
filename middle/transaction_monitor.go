package middle

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"hyperledger-fabric-copyright/conf"
)

// GetAllTransactions 获取区块链上的所有交易记录
func GetAllTransactions(_ context.Context, c *app.RequestContext) {
	log.Println("获取所有交易记录")

	// 确保BasicContract已初始化
	if conf.BasicContract == nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"error": "基础合约未初始化",
		})
		return
	}

	// 调用区块链上的GetAllCreatetranses函数
	log.Println("--> 执行交易: GetAllCreatetranses, 返回所有交易")
	evaluateResult, err := conf.BasicContract.EvaluateTransaction("GetAllCreatetranses")
	if err != nil {
		log.Printf("执行交易失败: %v", err)
		c.JSON(http.StatusInternalServerError, utils.H{
			"error": fmt.Sprintf("获取交易记录失败: %v", err),
		})
		return
	}

	// 直接将区块链返回的JSON数据发送给客户端
	c.Data(http.StatusOK, "application/json", evaluateResult)
}
