package conf

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
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
	}
	err = yaml.Unmarshal(dataBytes, &temp)
	Con.Mysql = temp.Mysql
	Con.Jwtkey = []byte(temp.Jwtkey)
	if err != nil {
		log.Fatal("解析 yaml 文件失败：", err)
	}
}
