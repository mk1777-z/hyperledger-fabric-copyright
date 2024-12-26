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

	_, err := contract.SubmitTransaction("CreateCreatetrans", trans.ID, trans.Name, trans.Seller, trans.Purchaser, trans.Purchaser, strconv.FormatFloat(trans.Price, 'f', -1, 64), trans.Transtime)
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
	defer db.Close() // 确保数据库连接在结束时关闭
	rows, err := db.Query("SELECT price, owner,transID FROM item WHERE name = ?", name.Name)
	if err != nil {
		log.Fatal("Query DataBase err:", err)
	}
	var price float64
	var seller string
	var transID string
	rows.Scan(&price, &seller, &transID)
	trans := conf.Createtrans{ID: assetId, Name: name.Name, Seller: seller, Purchaser: claims.Username, Price: price, Transtime: now.Format("Basic short date")}
	createAsset(conf.Contract, trans)
	transID = transID + " " + assetId
	_, err = db.Exec("UPDATE item SET owner=? transID =? WHERE name = ?", claims.Username, transID, name.Name)
	if err != nil {
		log.Fatal("Update DataBase err:", err)
	}
	// 返回结果
	c.JSON(http.StatusOK, utils.H{
		"message": "Transaction successful",
	})
}
