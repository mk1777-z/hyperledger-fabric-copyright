package middle

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"gopkg.in/yaml.v3"
)

// 华为云OBS凭证配置
type ObsCredentials struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

// 华为云相关配置
type HuaweiConfig struct {
	Obs ObsCredentials `yaml:"obs"`
}

// OBS配置变量
var (
	obsEndPoint string = "https://obs.cn-east-3.myhuaweicloud.com"
	obsBucket   string = "huaweibucket-48f4"
	obsClient   *obs.ObsClient
	initObsErr  error
)

// 初始化OBS客户端
func InitializeOBSClient() {
	configPath := "../config.yaml"

	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		initObsErr = fmt.Errorf("读取配置文件失败 '%s': %w", configPath, err)
		log.Printf("初始化OBS客户端错误: %v", initObsErr)
		return
	}

	var config HuaweiConfig
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		initObsErr = fmt.Errorf("解析配置文件失败 '%s': %w", configPath, err)
		log.Printf("初始化OBS客户端错误: %v", initObsErr)
		return
	}

	ak := config.Obs.AccessKeyID
	sk := config.Obs.SecretAccessKey

	if ak == "" || sk == "" {
		initObsErr = fmt.Errorf("OBS访问密钥缺失，配置文件 '%s'", configPath)
		log.Printf("初始化OBS客户端错误: %v", initObsErr)
		return
	}

	obsClient, initObsErr = obs.New(ak, sk, obsEndPoint)
	if initObsErr != nil {
		log.Printf("创建OBS客户端实例错误: %v", initObsErr)
	}
}

func Upload(_ context.Context, c *app.RequestContext) {
	var uploadInfo conf.Upload
	if err := c.Bind(&uploadInfo); err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效请求体", "error": err.Error()})
		log.Printf("请求体绑定错误: %v", err)
		return
	}

	InitializeOBSClient()
	// 检查OBS客户端是否成功初始化
	if obsClient == nil || initObsErr != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "图片存储服务不可用"})
		return
	}

	// JWT认证
	tokenString := c.GetHeader("Authorization")
	if tokenString == nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "缺少授权令牌"})
		return
	}
	if !strings.HasPrefix(string(tokenString), "Bearer ") {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效的授权头格式"})
		return
	}
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", t.Header["alg"])
		}
		return []byte(conf.Con.Jwtkey), nil
	})

	if err != nil || !token.Valid {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				c.Status(http.StatusUnauthorized)
				c.JSON(http.StatusUnauthorized, utils.H{"message": "令牌已过期"})
				return
			}
		}
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "无效令牌", "error": err.Error()})
		return
	}
	claims := token.Claims.(*conf.UserClaims)

	// 图片处理
	if uploadInfo.Img == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "缺少图片数据"})
		return
	}

	// 解码Base64图片
	base64Image := uploadInfo.Img
	if commaIndex := strings.Index(base64Image, ","); commaIndex != -1 {
		base64Image = base64Image[commaIndex+1:]
	}
	imageData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "无效的Base64图片数据", "error": err.Error()})
		return
	}
	if len(imageData) == 0 {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "解码后的图片数据为空"})
		return
	}

	// 上传到OBS
	objectKey := strconv.Itoa(uploadInfo.ID)
	input := &obs.PutObjectInput{}
	input.Bucket = obsBucket
	input.Key = objectKey
	input.Body = bytes.NewReader(imageData)
	_, err = obsClient.PutObject(input)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "上传图片到存储失败", "error": err.Error()})
		return
	}

	// 构建公共OBS URL
	obsDomain := strings.TrimPrefix(obsEndPoint, "https://")
	imageURL := fmt.Sprintf("https://%s.%s/%s", obsBucket, obsDomain, objectKey)

	// 数据库交互
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库连接错误"})
		return
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "数据库连接错误"})
		return
	}

	startTime := time.Now()
	assetID := fmt.Sprintf("asset%d", startTime.UnixNano()/1e6)

	_, err = db.Exec(
		"INSERT INTO item (id, name, owner, simple_dsc, dsc, price, img, on_sale, start_time, transID, category) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uploadInfo.ID,
		uploadInfo.Name,
		claims.Username,
		uploadInfo.Simple_dsc,
		uploadInfo.Dsc,
		uploadInfo.Price,
		imageURL,
		uploadInfo.On_sale,
		startTime,
		assetID,
		uploadInfo.Category,
	)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "保存项目详情失败", "error": err.Error()})
		return
	}

	// 区块链交易
	_, err = conf.BasicContract.SubmitTransaction(
		"CreateCreatetrans",
		assetID,
		uploadInfo.Name,
		"admin",
		claims.Username,
		"0",
		startTime.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": "项目已保存并上传图片，但区块链记录失败",
			"assetID": assetID,
			"itemID":  uploadInfo.ID,
			"error":   err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
	c.JSON(http.StatusOK, utils.H{"message": "创建项目成功", "assetID": assetID, "itemID": uploadInfo.ID, "imageURL": imageURL})
}
