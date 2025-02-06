package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func createAsset(contract *client.Contract, trans conf.Createtrans) {
	fmt.Printf("\n--> Submit Transaction<-- \n")

	_, err := contract.SubmitTransaction("CreateCreatetrans", trans.ID, trans.Name, trans.Seller, trans.Purchaser, strconv.FormatFloat(trans.Price, 'f', -1, 64), trans.Transtime)
	if err != nil {
		log.Fatal("failed to submit transaction: %w", err)
		return
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

func Transaction(_ context.Context, c *app.RequestContext) {
	type temp struct {
		Name string
	}
	var name temp
	c.Bind(&name)
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

	now := time.Now()
	assetId := fmt.Sprintf("asset%d", now.Unix()*1e3+int64(now.Nanosecond())/1e6)
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT price, owner, transID FROM item WHERE name = ?", name.Name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database query error"})
		return
	}
	defer rows.Close()

	var price float64
	var seller string
	var transID *string
	for rows.Next() {
		if err := rows.Scan(&price, &seller, &transID); err != nil {
			c.Status(http.StatusInternalServerError)
			c.JSON(http.StatusInternalServerError, utils.H{"message": "Failed to scan database row"})
			return
		}
	}
	if rows.Err() != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Error during row iteration"})
		return
	}

	if seller == claims.Username {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "不可购买自己的版权",
		})
		return // 提前返回，避免后续操作
	}

	// 新增：调用资金链 Transfer 函数
	_, err = conf.FundsContract.SubmitTransaction(
		"Transfer",
		claims.Username, // from（买家）
		seller,          // to（卖家）
		strconv.FormatFloat(price, 'f', -1, 64),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "余额不足，请先至个人中心充值",
		})
		return
	}

	trans := conf.Createtrans{ID: assetId, Name: name.Name, Seller: seller, Purchaser: claims.Username, Price: price, Transtime: time.Now().Format("2006-01-02 15:04:05")}
	createAsset(conf.BasicContract, trans)

	if transID == nil {
		transID = new(string)
		*transID = assetId
	} else {
		*transID = *transID + " " + assetId
	}

	_, err = db.Exec("UPDATE item SET owner=?, transID =? WHERE name = ?", claims.Username, &transID, name.Name)
	if err != nil {
		log.Fatal("Update DataBase err:", err)
	}
	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"message": "Transaction successful",
	})
}
