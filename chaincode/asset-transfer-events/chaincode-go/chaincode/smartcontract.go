package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	Owner   string  `json:"owner"`   // 使用用户ID作为唯一标识
	Balance float64 `json:"balance"` //用户账户余额
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, owner string, balance float64) error {
	existing, err := s.readState(ctx, owner)
	if err == nil && existing != nil {
		return fmt.Errorf("the asset %s already exists", owner)
	}

	asset := Asset{
		Owner:   owner,
		Balance: balance,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	ctx.GetStub().SetEvent("CreateAsset", assetJSON)
	return ctx.GetStub().PutState(owner, assetJSON)
}

// readstate只在readasset时被调用，不能被外部其他用户调用查询,小写代表为私有函数
func (s *SmartContract) readState(ctx contractapi.TransactionContextInterface, owner string) ([]byte, error) {
	assetJSON, err := ctx.GetStub().GetState(owner)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", owner)
	}

	return assetJSON, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, owner string) (*Asset, error) {
	assetJSON, err := s.readState(ctx, owner)
	if err != nil {
		return nil, err
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// 充值
func (s *SmartContract) TopUp(ctx contractapi.TransactionContextInterface, owner string, amount float64) error {
	// 1. 获取现有资产
	asset, err := s.ReadAsset(ctx, owner)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}

	// 2. 修改余额
	asset.Balance += amount

	// 3. 更新状态
	return s.updateAsset(ctx, owner, asset.Balance)
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) updateAsset(ctx contractapi.TransactionContextInterface, owner string, balance float64) error {
	_, err := s.readState(ctx, owner)
	if err != nil {
		return err
	}

	// overwriting original asset with new asset
	asset := Asset{
		Owner:   owner,
		Balance: balance,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	ctx.GetStub().SetEvent("UpdateAsset", assetJSON)
	return ctx.GetStub().PutState(owner, assetJSON)
}

// Withdraw 提现功能（新增）
func (s *SmartContract) Withdraw(ctx contractapi.TransactionContextInterface, owner string, amount float64) error {
	asset, err := s.ReadAsset(ctx, owner)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}

	if asset.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	asset.Balance -= amount
	return s.updateAsset(ctx, owner, asset.Balance)
}

// Transfer 转账功能（新增）
func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, from string, to string, amount float64) error {
	// 转出方操作
	fromAsset, err := s.ReadAsset(ctx, from)
	if err != nil {
		return fmt.Errorf("sender error: %v", err)
	}
	if fromAsset.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}
	fromAsset.Balance -= amount

	// 转入方操作
	toAsset, err := s.ReadAsset(ctx, to)
	if err != nil {
		return fmt.Errorf("recipient error: %v", err)
	}
	toAsset.Balance += amount

	// 原子更新双方状态
	if err := s.updateAsset(ctx, from, fromAsset.Balance); err != nil {
		return err
	}
	if err := s.updateAsset(ctx, to, toAsset.Balance); err != nil {
		// 回滚转出操作
		fromAsset.Balance += amount
		_ = s.updateAsset(ctx, from, fromAsset.Balance)
		return err
	}

	// 触发转账事件
	transferEvent := map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
	}
	eventJSON, _ := json.Marshal(transferEvent)
	ctx.GetStub().SetEvent("Transfer", eventJSON)
	return nil
}
