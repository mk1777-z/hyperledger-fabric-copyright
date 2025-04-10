package conf

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// Org2 (监管者) 连接参数
	regulatorMspID        = "Org2MSP"
	regulatorCryptoPath   = "/home/sample/fabric1/scripts/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com"
	regulatorCertPath     = regulatorCryptoPath + "/users/User1@org2.example.com/msp/signcerts"
	regulatorKeyPath      = regulatorCryptoPath + "/users/User1@org2.example.com/msp/keystore"
	regulatorTlsCertPath  = regulatorCryptoPath + "/peers/peer0.org2.example.com/tls/ca.crt"
	regulatorPeerEndpoint = "dns:///localhost:9051"
	regulatorGatewayPeer  = "peer0.org2.example.com"
)

var (
	// 监管者专用合约
	RegulatorContractOrg2 *client.Contract
)

// 初始化监管者区块链连接
func InitRegulatorConnection() error {
	log.Println("正在初始化监管者专用区块链连接...")

	// 创建监管者客户端连接
	clientConnection, err := newRegulatorGrpcConnection()
	if err != nil {
		log.Printf("创建监管者gRPC连接失败: %v", err)
		return err
	}

	// 获取监管者身份
	id, err := newRegulatorIdentity()
	if err != nil {
		log.Printf("获取监管者身份失败: %v", err)
		return err
	}

	// 获取监管者签名
	sign, err := newRegulatorSign()
	if err != nil {
		log.Printf("获取监管者签名失败: %v", err)
		return err
	}

	// 创建监管者网关连接
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
		log.Printf("创建监管者网关连接失败: %v", err)
		return err
	}

	// 获取网络并初始化监管者合约
	network := gw.GetNetwork(channelName)
	RegulatorContractOrg2 = network.GetContract(regulatorChaincodeName)

	log.Println("监管者区块链连接初始化成功")
	return nil
}

// 创建监管者gRPC连接
func newRegulatorGrpcConnection() (*grpc.ClientConn, error) {
	certificate, err := loadRegulatorCertificate()
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, regulatorGatewayPeer)

	connection, err := grpc.Dial(regulatorPeerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("监管者节点连接失败: %w", err)
	}

	return connection, nil
}

// 加载监管者TLS证书
func loadRegulatorCertificate() (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(regulatorTlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("读取监管者证书文件失败: %w", err)
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, fmt.Errorf("解析监管者证书失败: %w", err)
	}

	return certificate, nil
}

// 创建监管者身份实例
func newRegulatorIdentity() (*identity.X509Identity, error) {
	certificate, err := loadRegulatorX509Certificate()
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity(regulatorMspID, certificate)
}

// 加载监管者X509证书
func loadRegulatorX509Certificate() (*x509.Certificate, error) {
	files, err := os.ReadDir(regulatorCertPath)
	if err != nil {
		return nil, fmt.Errorf("读取监管者证书目录失败: %w", err)
	}

	if len(files) < 1 {
		return nil, fmt.Errorf("监管者证书文件未找到")
	}

	certificatePEM, err := os.ReadFile(regulatorCertPath + "/" + files[0].Name())
	if err != nil {
		return nil, fmt.Errorf("读取监管者证书文件失败: %w", err)
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, fmt.Errorf("解析监管者证书失败: %w", err)
	}

	return certificate, nil
}

// 创建监管者签名实例
func newRegulatorSign() (identity.Sign, error) {
	files, err := os.ReadDir(regulatorKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取监管者私钥目录失败: %w", err)
	}

	if len(files) < 1 {
		return nil, fmt.Errorf("监管者私钥文件未找到")
	}

	privateKeyPEM, err := os.ReadFile(regulatorKeyPath + "/" + files[0].Name())
	if err != nil {
		return nil, fmt.Errorf("读取监管者私钥文件失败: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析监管者私钥失败: %w", err)
	}

	return identity.NewPrivateKeySign(privateKey)
}
