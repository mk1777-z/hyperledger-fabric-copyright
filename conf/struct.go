package conf

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

type Mysql struct {
	DbUser     string `yaml:"dbUser"`
	DbPassword string `yaml:"dbPassword"`
	DbName     string `yaml:"dbName"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Mysql  Mysql  `yaml:"mysql"`
	Jwtkey []byte `yaml:"jwtkey"`
}

type Upload struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Simple_dsc string  `json:"simple_dsc"`
	Dsc        string  `json:"dsc"`
	Price      float32 `json:"price"`
	Img        string  `json:"img"`
	On_sale    bool    `json:"on_sale"`
}

type UpdateItem struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	Dsc         string  `json:"dsc"`
	Sale        bool    `json:"on_sale"`
}
type SimpleContract struct {
	contractapi.Contract
}
