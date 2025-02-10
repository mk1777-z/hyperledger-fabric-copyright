package middle

import (
	"context"
	"encoding/json"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// 初始化账户余额
func initAccount(contract *client.Contract, owner string, balance float64) {
	fmt.Printf("\n--> Submit Transaction: InitAccount <--\n")

	_, err := contract.SubmitTransaction("CreateAsset", owner, strconv.FormatFloat(balance, 'f', -1, 64))
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
		return
	}

	fmt.Printf("*** Account initialized successfully: Owner=%s, Initial Balance=%.2f\n", owner, balance)
}

// 新增查询函数
func queryBalance(contract *client.Contract, owner string) (float64, error) {
	fmt.Printf("\n--> Evaluate Transaction: QueryBalance <--\n")

	result, err := contract.EvaluateTransaction("ReadAsset", owner)
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate transaction: %v", err)
	}

	var asset conf.Asset
	if err := json.Unmarshal(result, &asset); err != nil {
		return 0, fmt.Errorf("failed to parse asset data: %v", err)
	}

	return asset.Balance, nil
}

// 提现
func withdraw(contract *client.Contract, owner string, amount float64) {
	fmt.Printf("\n--> Submit Transaction: Withdraw <--\n")

	_, err := contract.SubmitTransaction("Withdraw", owner, strconv.FormatFloat(amount, 'f', -1, 64))
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
		return
	}

	fmt.Printf("*** Withdraw successful: Owner=%s, Amount=%.2f\n", owner, amount)
}

// 充值
// func topUp(contract *client.Contract, owner string, amount float64) {
// 	fmt.Printf("\n--> Submit Transaction: TopUp <--\n")

//		// 尝试读取现有账户
//		_, err := contract.EvaluateTransaction("ReadAsset", owner)
//		if err != nil {
//			// 如果不存在则初始化
//			initAccount(contract, owner, amount)
//		} else {
//			// 存在则充值
//			_, err = contract.SubmitTransaction("TopUp", owner, strconv.FormatFloat(amount, 'f', -1, 64))
//			if err != nil {
//				log.Fatalf("failed to submit transaction: %v", err)
//			}
//		}
//		fmt.Printf("*** TopUp successful: Owner=%s, Amount=%.2f\n", owner, amount)
//	}
func topUp(contract *client.Contract, owner string, amount float64) {
	fmt.Printf("\n--> Submit Transaction: TopUp <--\n")

	// 直接充值（不再检查账户是否存在）
	_, err := contract.SubmitTransaction("TopUp", owner, strconv.FormatFloat(amount, 'f', -1, 64))
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
		return
	}

	fmt.Printf("*** TopUp successful: Owner=%s, Amount=%.2f\n", owner, amount)
}

// 转账
func transfer(contract *client.Contract, from string, to string, amount float64) {
	fmt.Printf("\n--> Submit Transaction: Transfer <--\n")

	_, err := contract.SubmitTransaction("Transfer", from, to, strconv.FormatFloat(amount, 'f', -1, 64))
	if err != nil {
		log.Fatalf("failed to submit transaction: %v", err)
		return
	}

	fmt.Printf("*** Transfer successful: From=%s, To=%s, Amount=%.2f\n", from, to, amount)
}

// 处理用户请求
func HandleAccount(_ context.Context, c *app.RequestContext) {
	type Request struct {
		Action     string  `json:"action"`      // 操作类型：init, withdraw, topup, transfer, query
		Owner      string  `json:"owner"`       // 账户所有者
		To         string  `json:"to"`          // 转账目标账户（仅用于 transfer 操作）
		Amount     float64 `json:"amount"`      // 金额
		Balance    float64 `json:"balance"`     // 初始余额（仅用于 init 操作）
		QueryOwner string  `json:"query_owner"` // 查询目标账户（可选，用于 query）
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid request body"})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 返回 JWT 密钥
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		// 如果 token 无效，返回 401 未授权错误
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}
	username := claims.Username

	// 根据操作类型校验权限
	switch req.Action {
	case "init":
		// 初始化账户只能操作自己
		if req.Owner != username {
			c.Status(http.StatusForbidden)
			c.JSON(http.StatusForbidden, utils.H{"message": "Cannot initialize account for others"})
			return
		}
		initAccount(conf.FundsContract, req.Owner, req.Balance)
		c.JSON(http.StatusOK, utils.H{"message": "Account initialized successfully"})

	case "withdraw", "topup":
		// 充值和提现只能操作自己
		if req.Owner != username {
			c.Status(http.StatusForbidden)
			c.JSON(http.StatusForbidden, utils.H{"message": "Cannot modify other accounts"})
			return
		}
		if req.Amount <= 0 {
			c.Status(http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, utils.H{"message": "金额必须大于0"})
			return
		}
		if req.Action == "withdraw" {
			withdraw(conf.FundsContract, req.Owner, req.Amount)
		} else {
			topUp(conf.FundsContract, req.Owner, req.Amount)
		}
		c.JSON(http.StatusOK, utils.H{"message": "Operation successful"})

	case "transfer":
		// 转账发起方必须是本人
		if req.Owner != username {
			c.Status(http.StatusForbidden)
			c.JSON(http.StatusForbidden, utils.H{"message": "Cannot transfer from others"})
			return
		}
		transfer(conf.FundsContract, req.Owner, req.To, req.Amount)
		c.JSON(http.StatusOK, utils.H{"message": "Transfer successful"})

	case "query":
		// // 查询目标账户处理
		// targetOwner := req.Owner
		// if req.QueryOwner != "" {
		// 	targetOwner = req.QueryOwner
		// }

		// // 只能查询自己账户（除非有管理员权限）
		// if targetOwner != username {
		// 	c.Status(http.StatusForbidden)
		// 	c.JSON(http.StatusForbidden, utils.H{"message": "Cannot query other accounts"})
		// 	return
		// }

		// balance, err := queryBalance(conf.FundsContract, targetOwner)
		// if err != nil {
		// 	c.Status(http.StatusInternalServerError)
		// 	c.JSON(http.StatusInternalServerError, utils.H{"message": err.Error()})
		// 	return
		// }
		// c.JSON(http.StatusOK, utils.H{"owner": targetOwner, "balance": balance})
		balance, err := queryBalance(conf.FundsContract, username)
		if err != nil {
			// 如果账户不存在返回0
			if strings.Contains(err.Error(), "does not exist") {
				c.JSON(http.StatusOK, utils.H{
					"owner":   username,
					"balance": 0.0,
				})
				return
			}
			c.Status(http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, utils.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, utils.H{
			"owner":   username,
			"balance": balance,
		})

	default:
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Invalid action type"})
	}
}
