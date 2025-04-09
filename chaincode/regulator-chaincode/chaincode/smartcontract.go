package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract 提供了主合约结构体
type SmartContract struct {
	contractapi.Contract
}

// AuditRecord 审核记录
type AuditRecord struct {
	TradeID   string `json:"tradeId"`
	Decision  string `json:"decision"` // APPROVE/REJECT
	Comment   string `json:"comment"`
	Regulator string `json:"regulator"`
	Timestamp int64  `json:"timestamp"`
}

// AuditTrade 执行审核操作
func (s *SmartContract) AuditTrade(ctx contractapi.TransactionContextInterface,
	tradeID string,
	decision string,
	comment string) error {

	// 获取交易状态
	tradeData, err := ctx.GetStub().GetState(tradeID)
	if err != nil || tradeData == nil {
		return fmt.Errorf("trade %s not found", tradeID)
	}

	var trade map[string]interface{}
	json.Unmarshal(tradeData, &trade)

	// 记录审核
	record := AuditRecord{
		TradeID:   tradeID,
		Decision:  decision,
		Comment:   comment,
		Regulator: getCallerID(ctx),
		Timestamp: time.Now().Unix(),
	}

	// 使用CreateCompositeKey创建复合键，而不是字符串拼接
	// 使用"AUDIT"作为前缀，确保与GetAuditHistory中的一致
	recordKey, err := ctx.GetStub().CreateCompositeKey("AUDIT", []string{tradeID})
	if err != nil {
		return fmt.Errorf("创建复合键失败: %v", err)
	}

	recordData, _ := json.Marshal(record)
	if err := ctx.GetStub().PutState(recordKey, recordData); err != nil {
		return err
	}

	// 直接在当前合约内更新交易状态
	trade["status"] = decision
	trade["lastUpdated"] = time.Now().Unix()

	updatedTradeData, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("failed to marshal updated trade data: %v", err)
	}

	if err := ctx.GetStub().PutState(tradeID, updatedTradeData); err != nil {
		return fmt.Errorf("failed to update trade status: %v", err)
	}

	return nil
}

// GetAuditHistory 获取审核记录
func (s *SmartContract) GetAuditHistory(ctx contractapi.TransactionContextInterface, tradeID string) ([]AuditRecord, error) {
	// 修复部分复合键的错误使用
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("AUDIT", []string{tradeID})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []AuditRecord
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var record AuditRecord
		json.Unmarshal(kv.Value, &record)
		records = append(records, record)
	}
	return records, nil
}

// CreateTrade 创建一个新的交易记录
func (s *SmartContract) CreateTrade(ctx contractapi.TransactionContextInterface,
	tradeID string, tradeJSON string) error {

	// 检查交易是否已存在
	exists, err := s.TradeExists(ctx, tradeID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("交易ID已存在: %s", tradeID)
	}

	// 直接存储交易数据
	return ctx.GetStub().PutState(tradeID, []byte(tradeJSON))
}

// TradeExists 检查交易是否存在
func (s *SmartContract) TradeExists(ctx contractapi.TransactionContextInterface,
	tradeID string) (bool, error) {

	tradeBytes, err := ctx.GetStub().GetState(tradeID)
	if err != nil {
		return false, fmt.Errorf("读取交易数据失败: %v", err)
	}

	return tradeBytes != nil, nil
}

// 获取调用者身份
func getCallerID(ctx contractapi.TransactionContextInterface) string {
	id, _ := ctx.GetClientIdentity().GetID()
	return id
}
