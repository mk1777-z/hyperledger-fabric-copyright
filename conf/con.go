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
)

var Contract *client.Contract
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
	}
	err = yaml.Unmarshal(dataBytes, &temp)
	Con.Mysql = temp.Mysql
	Con.Jwtkey = []byte(temp.Jwtkey)
	if err != nil {
		log.Fatal("解析 yaml 文件失败：", err)
	}

	clientConnection := newGrpcConnection()
	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "basic"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	Initcontract := network.GetContract(chaincodeName)
	Contract = Initcontract
}
