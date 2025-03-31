package middle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// AuditRecord 审核记录结构，与智能合约保持一致
type AuditRecord struct {
	TradeID   string `json:"tradeId"`
	Decision  string `json:"decision"` // APPROVE/REJECT
	Comment   string `json:"comment"`
	Regulator string `json:"regulator"`
	Timestamp int64  `json:"timestamp"`
}

// 验证和解析交易ID
func validateAndParseTradeID(tradeID string) (string, error) {
	// 如果传入的是多个交易ID（空格分隔），取最后一个作为默认值
	ids := strings.Split(tradeID, " ")
	if len(ids) == 0 {
		return "", fmt.Errorf("无效的交易ID格式")
	}

	// 验证交易ID格式是否正确（假设格式为 "asset" 开头）
	validID := ids[len(ids)-1] // 默认取最后一个
	if !strings.HasPrefix(validID, "asset") {
		return "", fmt.Errorf("交易ID格式错误，必须以'asset'开头")
	}

	return validID, nil
}

// 验证交易是否存在
func verifyTradeExists(contract *client.Contract, tradeID string) (bool, error) {
	fmt.Printf("\n--> Evaluate Transaction: Verify Trade Exists for ID: %s <-- \n", tradeID)

	// 使用与information.go中相同的方法读取交易信息
	result, err := contract.EvaluateTransaction("ReadCreatetrans", tradeID)
	if err != nil {
		log.Printf("*** 交易验证失败: ID=%s, 错误=%v ***", tradeID, err)
		// 如果交易不存在，通常会返回特定错误
		if strings.Contains(err.Error(), "交易不存在") || strings.Contains(err.Error(), "does not exist") {
			return false, fmt.Errorf("交易 ID %s 不存在", tradeID)
		}
		return false, err
	}

	// 如果结果为空，表示交易不存在
	if len(result) == 0 {
		log.Printf("*** 交易 %s 返回空结果 ***", tradeID)
		return false, fmt.Errorf("交易 ID %s 返回空结果", tradeID)
	}

	fmt.Printf("*** 获取到交易数据: %s ***\n", string(result))

	// 验证结果是否为有效的JSON
	var transactionDetails map[string]interface{}
	if err := json.Unmarshal(result, &transactionDetails); err != nil {
		log.Printf("*** 交易数据解析失败: ID=%s, 错误=%v ***", tradeID, err)
		return false, fmt.Errorf("交易数据格式无效: %v", err)
	}

	// 检查必要字段是否存在
	if _, ok := transactionDetails["ID"]; !ok {
		log.Printf("*** 交易数据缺少必要字段: ID=%s ***", tradeID)
		return false, fmt.Errorf("交易数据缺少必要字段")
	}

	fmt.Printf("*** 交易 %s 验证成功存在 ***\n", tradeID)
	return true, nil
}

// 创建审核记录
func createAuditRecord(contract *client.Contract, tradeID string, decision string, comment string) error {
	fmt.Printf("\n--> Submit Transaction: AuditTrade <-- \n")
	fmt.Printf("参数: tradeID=%s, decision=%s, comment=%s\n", tradeID, decision, comment)

	// 添加详细调试信息
	fmt.Printf("*** 合约对象信息: %#v ***\n", contract)
	fmt.Printf("*** 合约ID: %v ***\n", contract.ChaincodeName())

	// 检查合约是否为nil
	if contract == nil {
		return fmt.Errorf("监管合约对象为空，请检查初始化")
	}

	// 尝试先调用无害查询操作测试合约连接
	try := func() error {
		fmt.Println("测试合约连接...")
		_, err := contract.EvaluateTransaction("GetAuditHistory", tradeID)
		if err != nil {
			fmt.Printf("测试查询失败: %v\n", err)
			return err
		}
		fmt.Println("测试查询成功")
		return nil
	}

	if err := try(); err != nil {
		fmt.Printf("警告: 合约连接测试失败，但仍将尝试提交审核\n")
	}

	// 直接提交审核交易，添加更详细的错误处理
	result, err := contract.SubmitTransaction("AuditTrade", tradeID, decision, comment)
	if err != nil {
		log.Printf("*** 审核交易提交失败: %v ***", err)

		// 错误分类处理
		errMsg := err.Error()
		if strings.Contains(errMsg, "regulator") {
			return fmt.Errorf("跨链码调用失败: 未找到 regulator 链码，请确认链码已正确部署")
		} else if strings.Contains(errMsg, "trade") && strings.Contains(errMsg, "not found") {
			return fmt.Errorf("交易ID %s 在监管合约中未找到", tradeID)
		} else if strings.Contains(errMsg, "access denied") || strings.Contains(errMsg, "permission") {
			return fmt.Errorf("权限错误: 调用者可能没有执行此操作的权限")
		} else if strings.Contains(errMsg, "endorsement") {
			return fmt.Errorf("背书失败: 可能是跨链码配置问题或通道配置错误")
		}

		return fmt.Errorf("审核交易失败: %v", err)
	}

	fmt.Printf("*** 审核记录提交成功，结果: %s ***\n", string(result))
	return nil
}

// 获取审核历史记录
func getAuditHistory(contract *client.Contract, tradeID string) ([]AuditRecord, error) {
	fmt.Printf("\n--> Evaluate Transaction: GetAuditHistory <-- \n")

	result, err := contract.EvaluateTransaction("GetAuditHistory", tradeID)
	if err != nil {
		log.Printf("failed to evaluate transaction: %v", err)
		return nil, err
	}

	var records []AuditRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		log.Printf("failed to unmarshal audit history: %v", err)
		return nil, err
	}

	return records, nil
}

// 验证是否为监管者身份
func isRegulator(username, password string) bool {
	// 首先检查用户名是否为"监管者"
	if username != "监管者" {
		log.Printf("用户 %s 不是监管者", username)
		return false
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s",
		conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		return false
	}
	defer db.Close()

	// 查询数据库中的密码
	var dbPassword string
	err = db.QueryRow("SELECT password FROM user WHERE username = ?", username).Scan(&dbPassword)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
		return false
	}

	// 验证密码是否匹配
	if password != dbPassword {
		log.Printf("监管者密码验证失败")
		return false
	}

	log.Printf("监管者身份验证成功")
	return true
}

func AuditTrade(_ context.Context, c *app.RequestContext) {
	log.Println("AuditTrade API called.")

	// 定义请求结构
	type AuditRequest struct {
		TradeID  string `json:"tradeId"`
		Decision string `json:"decision"` // APPROVE/REJECT
		Comment  string `json:"comment"`
		Password string `json:"password"` // 用于验证监管者身份
	}

	// 解析请求
	var req AuditRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效的请求参数"})
		return
	}

	log.Printf("收到审核请求: 交易ID=%s, 决定=%s", req.TradeID, req.Decision)

	// 检查交易ID是否为空
	if req.TradeID == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "交易ID不能为空"})
		return
	}

	// 验证并解析交易ID
	validTradeID, err := validateAndParseTradeID(req.TradeID)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": err.Error()})
		return
	}

	// 更新请求中的交易ID为验证后的ID
	req.TradeID = validTradeID

	// 先验证交易是否存在
	exists, err := verifyTradeExists(conf.BasicContract, req.TradeID)
	if err != nil || !exists {
		statusMsg := "交易不存在或无效"
		if err != nil {
			statusMsg = err.Error()
		}
		log.Printf("*** 交易验证失败: %s ***", statusMsg)
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "审核失败",
			"error":   statusMsg,
		})
		return
	}

	// 检查RegulatorContract是否已正确初始化
	if conf.RegulatorContract == nil {
		log.Printf("*** 严重错误: RegulatorContract 未初始化 ***")
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "系统配置错误",
			"error":   "监管合约未正确初始化",
		})
		return
	}

	log.Printf("*** 交易验证通过，继续执行审核 ***")

	// 验证决定值是否有效
	if req.Decision != "APPROVE" && req.Decision != "REJECT" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "决定必须是 APPROVE 或 REJECT"})
		return
	}

	// 验证用户权限
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "缺少授权令牌"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌声明"})
		return
	}

	// 检查用户是否为监管者并验证密码
	if !isRegulator(claims.Username, req.Password) {
		c.Status(http.StatusForbidden)
		c.JSON(http.StatusForbidden, utils.H{"message": "您不是监管者或密码验证失败"})
		return
	}

	// 调用链码函数
	err = createAuditRecord(conf.RegulatorContract, req.TradeID, req.Decision, req.Comment)
	if err != nil {
		log.Printf("*** 审核失败: %v ***", err)
		// 区分不同类型的错误，返回更具体的错误信息
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "交易ID无效") {
			statusCode = http.StatusBadRequest
		}

		c.Status(statusCode)
		c.JSON(statusCode, utils.H{
			"message": "审核交易失败",
			"error":   errorMsg,
		})
		return
	}

	// 返回成功结果
	c.JSON(http.StatusOK, utils.H{
		"message": "审核成功",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
		"status":  req.Decision,
	})
}

// GetAuditHistory 获取交易的审核历史
func GetAuditHistory(_ context.Context, c *app.RequestContext) {
	log.Println("GetAuditHistory API called.")

	// 获取交易ID参数
	tradeID := c.Query("tradeId")
	if tradeID == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "缺少交易ID参数"})
		return
	}

	// 验证并解析交易ID
	validTradeID, err := validateAndParseTradeID(tradeID)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": err.Error()})
		return
	}

	// 使用验证后的交易ID
	tradeID = validTradeID

	// 验证用户权限
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "缺少授权令牌"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌"})
		return
	}

	// 验证 token 是否有效
	if _, ok := token.Claims.(*conf.UserClaims); !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌声明"})
		return
	}

	// 调用链码获取审核历史
	records, err := getAuditHistory(conf.RegulatorContract, tradeID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "获取审核历史失败",
			"error":   err.Error(),
		})
		return
	}

	// 返回审核历史记录
	c.JSON(http.StatusOK, utils.H{
		"records": records,
		"count":   len(records),
	})
}
