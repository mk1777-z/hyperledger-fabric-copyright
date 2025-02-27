package conf

import "github.com/dgrijalva/jwt-go"

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
	APIKey string `yaml:"APIKey"`
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

type Createtrans struct {
	ID        string
	Name      string
	Seller    string
	Purchaser string
	Price     float64
	Transtime string
}

type Asset struct {
	Owner   string
	Balance float64
}

// UserClaims 用于 JWT 的声明

type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}
