package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract provides functions for managing Createtrans
type SmartContract struct {
	contractapi.Contract
}

// Createtrans 描述一次交易信息
type Createtrans struct {
	ID        string
	Name      string
	Seller    string
	Purchaser string
	Price     float64
	Transtime string
}

// InitLedger 添加一组初始的交易数据
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Initializing the ledger with predefined transactions...")
	createtranses := []Createtrans{
		{
			ID:        "tx1",
			Name:      "交易1",
			Seller:    "Alice",
			Purchaser: "Bob",
			Price:     1000.0,
			Transtime: "2024-01-01 10:00:00",
		},
		{
			ID:        "tx2",
			Name:      "交易2",
			Seller:    "Carol",
			Purchaser: "Dan",
			Price:     2000.0,
			Transtime: "2024-01-02 11:30:00",
		},
	}

	for _, createtrans := range createtranses {
		createtransJSON, err := json.Marshal(createtrans)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(createtrans.ID, createtransJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
		fmt.Printf("Initialized transaction with ID: %s\n", createtrans.ID)
	}
	return nil
}

// CreateCreatetrans 在账本中新增一次交易
func (s *SmartContract) CreateCreatetrans(
	ctx contractapi.TransactionContextInterface,
	id string,
	name string,
	seller string,
	purchaser string,
	price string,
	transtime string,
) error {
	fmt.Printf("Creating new transaction with ID: %s\n", id)
	exists, err := s.CreatetransExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the transaction %s already exists", id)
	}

	priceNum, _ := strconv.ParseFloat(price, 64)
	createtrans := Createtrans{
		ID:        id,
		Name:      name,
		Seller:    seller,
		Purchaser: purchaser,
		Price:     priceNum,
		Transtime: transtime,
	}
	createtransJSON, err := json.Marshal(createtrans)
	if err != nil {
		return err
	}

	fmt.Println("Successfully marshaled transaction data")
	err = ctx.GetStub().PutState(id, createtransJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}
	fmt.Printf("Successfully created transaction %s\n", id)
	return nil
}

// ReadCreatetrans 根据 ID 读取对应交易信息
func (s *SmartContract) ReadCreatetrans(ctx contractapi.TransactionContextInterface, id string) (*Createtrans, error) {
	fmt.Printf("Reading transaction with ID: %s\n", id)
	createtransJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if createtransJSON == nil {
		return nil, fmt.Errorf("the transaction %s does not exist", id)
	}

	var createtrans Createtrans
	err = json.Unmarshal(createtransJSON, &createtrans)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Successfully read transaction: %v\n", createtrans)
	return &createtrans, nil
}

// UpdateCreatetrans 更新已有交易信息
func (s *SmartContract) UpdateCreatetrans(
	ctx contractapi.TransactionContextInterface,
	id string,
	name string,
	seller string,
	purchaser string,
	price float64,
	transtime string,
) error {
	fmt.Printf("Updating transaction with ID: %s\n", id)
	exists, err := s.CreatetransExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the transaction %s does not exist", id)
	}

	createtrans := Createtrans{
		ID:        id,
		Name:      name,
		Seller:    seller,
		Purchaser: purchaser,
		Price:     price,
		Transtime: transtime,
	}
	createtransJSON, err := json.Marshal(createtrans)
	if err != nil {
		return err
	}
	fmt.Println("Successfully marshaled updated transaction data")
	err = ctx.GetStub().PutState(id, createtransJSON)
	if err != nil {
		return fmt.Errorf("failed to update state: %v", err)
	}
	fmt.Printf("Successfully updated transaction %s\n", id)
	return nil
}

// DeleteCreatetrans 从账本中删除一笔交易
func (s *SmartContract) DeleteCreatetrans(ctx contractapi.TransactionContextInterface, id string) error {
	fmt.Printf("Deleting transaction with ID: %s\n", id)
	exists, err := s.CreatetransExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the transaction %s does not exist", id)
	}

	err = ctx.GetStub().DelState(id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %v", err)
	}
	fmt.Printf("Successfully deleted transaction %s\n", id)
	return nil
}

// CreatetransExists 判断指定 ID 的交易是否存在
func (s *SmartContract) CreatetransExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	fmt.Printf("Checking if transaction %s exists\n", id)
	createtransJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return createtransJSON != nil, nil
}

// GetAllCreatetranses 返回账本中所有交易
func (s *SmartContract) GetAllCreatetranses(ctx contractapi.TransactionContextInterface) ([]*Createtrans, error) {
	fmt.Println("Getting all transactions from ledger...")
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var createtranses []*Createtrans
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var createtrans Createtrans
		err = json.Unmarshal(queryResponse.Value, &createtrans)
		if err != nil {
			return nil, err
		}
		createtranses = append(createtranses, &createtrans)
	}

	fmt.Printf("Successfully retrieved all transactions: %d transactions found\n", len(createtranses))
	return createtranses, nil
}
