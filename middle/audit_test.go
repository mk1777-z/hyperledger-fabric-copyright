package middle

import (
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
)

// TestRegulatorContract 用于测试监管合约是否正常工作
func TestRegulatorContract() error {
	if conf.RegulatorContract == nil {
		return fmt.Errorf("监管合约未初始化")
	}

	// 尝试获取链码信息
	chaincodeName := conf.RegulatorContract.ChaincodeName()
	log.Printf("监管合约名: %s", chaincodeName)

	// 尝试执行简单查询
	tradeID := "asset1712275200000" // 使用已知存在的交易ID
	result, err := conf.RegulatorContract.EvaluateTransaction("GetAuditHistory", tradeID)
	if err != nil {
		log.Printf("测试查询失败: %v", err)
		return err
	}

	log.Printf("测试成功: %s", string(result))
	return nil
}
