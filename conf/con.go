package conf

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"gopkg.in/yaml.v3"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "/home/sample/fabric1/scripts/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "dns:///localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"

	basicChaincodeName     = "basic"     // 资产交易链码
	fundsChaincodeName     = "funds"     // 资金操作链码
	regulatorChaincodeName = "regulator" // 监管链码
	channelName            = "mychannel"
)

var (
	BasicContract     *client.Contract // 资产交易合约
	FundsContract     *client.Contract // 资金操作合约
	RegulatorContract *client.Contract // 监管合约（通过Org1节点访问）
)
var Con Config

func Init() {
	fmt.Println("正在执行Init")
	dataBytes, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("读取 yaml 文件失败：", err)
	}
	var temp struct {
		Mysql  Mysql  `yaml:"mysql"`
		Jwtkey string `yaml:"jwtkey"`
		APIKey string `yaml:"APIKey"`
	}
	err = yaml.Unmarshal(dataBytes, &temp)
	Con.Mysql = temp.Mysql
	Con.Jwtkey = []byte(temp.Jwtkey)
	Con.APIKey = temp.APIKey
	if err != nil {
		log.Fatal("解析 yaml 文件失败：", err)
	}

	// 初始化数据库连接
	log.Println("开始初始化数据库连接...")
	errDB := InitDB()
	if errDB != nil {
		log.Printf("数据库初始化失败: %v", errDB)
		// 我们不应该在这里panic，而是记录错误并继续
		// 这样即使数据库连接失败，其他功能仍可用
		log.Println("警告: 数据库连接失败，统计功能将不可用")
	} else {
		log.Println("数据库初始化成功")
	}

	// 初始化区块链连接（Org1节点）
	clientConnection := newGrpcConnection()
	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	// 初始化基本合约（通过Org1节点）
	network := gw.GetNetwork(channelName)
	BasicContract = network.GetContract(basicChaincodeName)
	FundsContract = network.GetContract(fundsChaincodeName)
	RegulatorContract = network.GetContract(regulatorChaincodeName) // 基本连接的监管合约（备用）

	// 初始化监管者专用连接（通过Org2节点）
	errReg := InitRegulatorConnection()
	if errReg != nil {
		log.Printf("监管者专用连接初始化失败: %v，将使用Org1节点作为备用", errReg)
	} else {
		log.Println("监管者专用连接初始化成功")
	}
}
