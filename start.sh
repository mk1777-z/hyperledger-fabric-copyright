cd /home/hyperledger-fabric-copyright

# 创建一个临时脚本来重置transID字段
cat > reset_transid.go << 'EOF'
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"gopkg.in/yaml.v3"
	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Mysql struct {
		DbUser     string `yaml:"dbUser"`
		DbPassword string `yaml:"dbPassword"`
		DbName     string `yaml:"dbName"`
	} `yaml:"mysql"`
}

func main() {
	// 读取配置文件
	log.Println("读取数据库配置...")
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", 
		config.Mysql.DbUser, config.Mysql.DbPassword, config.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()
	
	// 执行重置操作
	log.Println("重置item表中的transID字段...")
	result, err := db.Exec("UPDATE item SET transID = NULL")
	if err != nil {
		log.Fatalf("重置transID失败: %v", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("成功重置 %d 条记录的transID", rowsAffected)
}
EOF

# 使用Go程序重置transID字段
echo "重置数据库中所有item的transID字段..."
go run reset_transid.go
echo "transID字段重置完成"

# 继续执行网络和链码初始化
cd /home/sample/fabric1/scripts/fabric-samples/test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go/ -ccl go
./network.sh deployCC -ccn funds -ccp ../asset-transfer-events/chaincode-go/ -ccl go
#./network.sh deployCC -ccn regulator -ccp ../asset-secured-agreement/chaincode-go/ -ccl go
./network.sh deployCC -ccn regulator -ccp ../regulator-chaincode/ -ccl go
cd /home/hyperledger-fabric-copyright
go build -o migrating migrate/main.go
./migrating

go run scripts/init/init_accounts.go

if [ $0 == "1" ]; then
    bash startproject.sh
fi
