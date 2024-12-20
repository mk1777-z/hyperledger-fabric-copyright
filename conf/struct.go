package conf

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
	Name       string `json:"name"`
	Simple_dsc string `json:"simple_dsc"`
	Dsc        string `json:"dsc"`
	Price      string `json:"price"`
	Img        string `json:"img"`
}
