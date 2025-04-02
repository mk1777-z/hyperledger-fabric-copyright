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

// 检查交易是否存在
func checkTradeExists(contract *client.Contract, tradeID string) bool {
	log.Printf("检查交易是否存在: %s", tradeID)

	result, err := contract.EvaluateTransaction("TradeExists", tradeID)
	if err != nil {
		log.Printf("检查交易失败: %v", err)
		return false
	}

	return string(result) == "true"
}

// 创建交易记录 - 内部使用，不作为API暴露
func createTradeIfNeeded(contract *client.Contract, tradeID string) error {
	// 检查交易是否已存在
	if exists := checkTradeExists(contract, tradeID); exists {
		log.Printf("交易已存在，无需创建: %s", tradeID)
		return nil
	}

	log.Printf("交易不存在，开始创建: %s", tradeID)

	// 创建交易对象
	trade := map[string]interface{}{
		"id":          tradeID,
		"description": fmt.Sprintf("自动创建的交易 %s", tradeID),
		"status":      "PENDING",
		"createdAt":   time.Now().Unix(),
	}

	// 转换为JSON
	tradeJSON, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("创建交易JSON失败: %v", err)
	}

	// 调用智能合约的CreateTrade函数
	_, err = contract.SubmitTransaction("CreateTrade", tradeID, string(tradeJSON))
	if err != nil {
		log.Printf("创建交易失败: %v", err)
		return fmt.Errorf("创建交易失败: %v", err)
	}

	log.Printf("交易 %s 创建成功", tradeID)
	return nil
}

// 创建审核记录
func submitAuditTransaction(contract *client.Contract, tradeID string, decision string, comment string) error {
	log.Printf("--> 提交审核交易: ID=%s, Decision=%s", tradeID, decision)

	// 直接调用智能合约的审核功能
	_, err := contract.SubmitTransaction("AuditTrade", tradeID, decision, comment)
	if err != nil {
		log.Printf("审核交易提交失败: %v", err)
		return fmt.Errorf("审核交易失败: %v", err)
	}

	log.Printf("审核记录提交成功")
	return nil
}

// 获取审核历史记录
func getAuditHistory(contract *client.Contract, tradeID string) ([]conf.AuditRecord, error) {
	log.Printf("--> 获取审核历史: ID=%s", tradeID)

	result, err := contract.EvaluateTransaction("GetAuditHistory", tradeID)
	if err != nil {
		log.Printf("获取审核历史失败: %v", err)
		return nil, err
	}

	// 打印原始返回结果，帮助调试
	log.Printf("审核历史原始返回: %s", string(result))

	// 检查返回的数据是否为空
	if len(result) == 0 || string(result) == "null" || string(result) == "[]" {
		log.Printf("审核历史为空，返回空数组")
		return []conf.AuditRecord{}, nil
	}

	var records []conf.AuditRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		log.Printf("解析审核历史失败: %v", err)
		// 尝试解析单个记录
		var singleRecord conf.AuditRecord
		errSingle := json.Unmarshal(result, &singleRecord)
		if errSingle == nil {
			log.Printf("成功解析单条审核记录")
			return []conf.AuditRecord{singleRecord}, nil
		}

		return nil, err
	}

	log.Printf("成功解析审核历史，包含 %d 条记录", len(records))
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

// 从basic链上查询交易信息
func getTradeInfoFromBasic(tradeID string) (map[string]interface{}, error) {
	log.Printf("从基础链查询交易信息: %s", tradeID)

	// 确保BasicContract已初始化
	if conf.BasicContract == nil {
		return nil, fmt.Errorf("基础合约未初始化")
	}

	// 调用basic链上的查询函数 - 使用ReadCreatetrans而非GetTrade
	log.Printf("--> 执行交易: ReadCreatetrans, 返回资产属性")
	evaluateResult, err := conf.BasicContract.EvaluateTransaction("ReadCreatetrans", tradeID)
	if err != nil {
		log.Printf("执行交易失败: %v", err)
		return nil, fmt.Errorf("从基础链查询交易失败: %v", err)
	}

	if len(evaluateResult) == 0 {
		return nil, fmt.Errorf("未在基础链上找到交易信息")
	}

	// 解析JSON结果
	var tradeInfo map[string]interface{}
	if err := json.Unmarshal(evaluateResult, &tradeInfo); err != nil {
		log.Printf("解析交易信息失败: %v", err)
		return nil, fmt.Errorf("解析交易信息失败: %v", err)
	}

	log.Printf("成功获取交易信息: %s", tradeID)
	return tradeInfo, nil
}

// AuditTrade - 主要API接口，提交审核
func AuditTrade(_ context.Context, c *app.RequestContext) {
	log.Println("AuditTrade API 被调用")

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
		log.Printf("解析请求失败: %v", err)
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

	// 检查RegulatorContract是否已正确初始化
	if conf.RegulatorContract == nil {
		log.Printf("严重错误: RegulatorContract 未初始化")
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "系统配置错误",
			"error":   "监管合约未正确初始化",
		})
		return
	}

	// 检查交易是否存在，如果不存在则创建
	err = createTradeIfNeeded(conf.RegulatorContract, req.TradeID)
	if err != nil {
		log.Printf("创建交易失败: %v", err)
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "交易准备失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("交易准备完成，开始审核...")

	// 提交审核
	err = submitAuditTransaction(conf.RegulatorContract, req.TradeID, req.Decision, req.Comment)
	if err != nil {
		log.Printf("审核失败: %v", err)
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "审核交易失败",
			"error":   err.Error(),
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

// GetAuditHistory - 获取交易的审核历史
func GetAuditHistory(_ context.Context, c *app.RequestContext) {
	log.Println("GetAuditHistory API 被调用")

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

	// 检查交易是否存在
	if exists := checkTradeExists(conf.RegulatorContract, tradeID); !exists {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "交易不存在",
		})
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

	// 返回审核历史记录，即使是空数组
	c.JSON(http.StatusOK, utils.H{
		"records": records,
		"count":   len(records),
	})
}

// GetTradeInfoForAudit - 获取交易信息供审核使用
func GetTradeInfoForAudit(_ context.Context, c *app.RequestContext) {
	log.Println("GetTradeInfoForAudit API 被调用")

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
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效的令牌声明"})
		return
	}

	// 检查用户是否为监管者
	if claims.Username != "监管者" {
		c.Status(http.StatusForbidden)
		c.JSON(http.StatusForbidden, utils.H{"message": "只有监管者可以查看待审核交易信息"})
		return
	}

	// 从基础链上获取交易信息
	tradeInfo, err := getTradeInfoFromBasic(tradeID)
	if err != nil {
		log.Printf("获取交易信息失败: %v", err)
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "获取交易信息失败",
			"error":   err.Error(),
		})
		return
	}

	// 检查监管链上是否已存在该交易
	tradeExists := false
	if conf.RegulatorContract != nil {
		tradeExists = checkTradeExists(conf.RegulatorContract, tradeID)
	}

	// 从交易信息中获取项目名称
	itemName, ok := tradeInfo["Name"].(string)
	if !ok || itemName == "" {
		log.Printf("交易信息中未找到有效的项目名称")
		c.JSON(http.StatusOK, utils.H{
			"message":     "获取交易信息成功，但未找到相关项目信息",
			"tradeInfo":   tradeInfo,
			"tradeExists": tradeExists,
		})
		return
	}

	// 连接数据库获取项目详细信息
	var itemDetails map[string]interface{}
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s",
		conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("数据库连接失败: %v", err)
		c.JSON(http.StatusOK, utils.H{
			"message":     "获取交易信息成功，但数据库连接失败",
			"tradeInfo":   tradeInfo,
			"tradeExists": tradeExists,
		})
		return
	}
	defer db.Close()

	// 查询项目详情
	var id int
	var name, simple_des, dsc, owner, img, start_time string
	var price float32
	var transID sql.NullString // 使用NullString处理可能为NULL的值

	err = db.QueryRow("SELECT id, name, simple_dsc, price, dsc, owner, img, start_time, transID FROM item WHERE name = ?",
		itemName).Scan(&id, &name, &simple_des, &price, &dsc, &owner, &img, &start_time, &transID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("未找到名称为 %s 的项目", itemName)
			c.JSON(http.StatusOK, utils.H{
				"message":     "获取交易信息成功，但未找到相关项目记录",
				"tradeInfo":   tradeInfo,
				"tradeExists": tradeExists,
			})
			return
		}
		log.Printf("查询项目信息失败: %v", err)
		c.JSON(http.StatusOK, utils.H{
			"message":     "获取交易信息成功，但查询项目信息失败",
			"tradeInfo":   tradeInfo,
			"tradeExists": tradeExists,
			"queryError":  err.Error(),
		})
		return
	}

	// 构建项目详情
	itemDetails = map[string]interface{}{
		"id":         id,
		"name":       name,
		"simple_dsc": simple_des,
		"price":      price,
		"dsc":        dsc,
		"owner":      owner,
		"img":        img,
		"start_time": start_time,
	}

	// 如果transID有值，添加到详情中
	if transID.Valid && transID.String != "" {
		itemDetails["transID"] = transID.String
	}

	// 返回交易信息和项目信息
	c.JSON(http.StatusOK, utils.H{
		"message":     "获取交易和项目信息成功",
		"tradeInfo":   tradeInfo,
		"itemDetails": itemDetails,
		"tradeExists": tradeExists,
	})
}
