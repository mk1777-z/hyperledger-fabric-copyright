package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"database/sql"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// 数据库连接信息
	dbUser     = "root"
	dbPassword = "12345678"
	dbName     = "fabric"
)

type User struct {
	Username string
	Password string
}

// JWT secret key
var jwtKey = []byte("123")

// UserClaims 用于 JWT 的声明
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func display(_ context.Context, c *app.RequestContext) {
	username, _ := c.Get(string(jwtKey))
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer db.Close() // 确保关闭数据库连接
	row, err := db.Query("Select owner from project where owner = ?", username)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, utils.H{
		"projecct": row,
	})
}

func renderHTML(h *server.Hertz) {
	// 加载HTML模板文件
	h.LoadHTMLGlob("HTML/project/*")

	h.Static("/static", "./")

	// 默认根路径返回一个 JSON 响应
	h.GET("/homepage", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "homepage.html", utils.H{
			"title": "Home",
		})
	})

	// 渲染 signin 页面
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "signin.html", utils.H{
			"title": "Sign In",
		})
	})

	// 渲染 signup 页面
	h.GET("/signup", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "signup.html", utils.H{
			"title": "Sign Up",
		})
	})

	h.GET("/information", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "information.html", utils.H{
			"title": "Information",
		})
	})
	h.GET("/display", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "display.html", utils.H{
			"title": "Information",
		})
	})
	h.GET("/upload", func(ctx context.Context, c *app.RequestContext) {
		c.HTML(consts.StatusOK, "upload.html", utils.H{
			"title": "Information",
		})
	})
}

func register(ctx context.Context, c *app.RequestContext) {
	var user User
	if err := c.Bind(&user); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer db.Close() // 确保关闭数据库连接

	// 检查用户名是否存在
	rows, err := db.Query("SELECT username FROM user WHERE username = ?", user.Username)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() { // 如果存在该用户
		c.Status(http.StatusBadRequest) // 返回400错误，表示用户已存在
		return
	}

	// 插入新用户
	if _, err := db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", user.Username, user.Password); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(consts.StatusOK)
}

// login 函数
func login(_ context.Context, c *app.RequestContext) {
	var user User
	if err := c.Bind(&user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 查询用户
	var storedPassword string
	err = db.QueryRow("SELECT password FROM user WHERE username = ?", user.Username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(http.StatusUnauthorized) // 用户不存在
		} else {
			c.Status(http.StatusInternalServerError) // 其他错误
		}
		return
	}

	if user.Password != storedPassword {
		c.Status(http.StatusUnauthorized) // 密码不正确
		return
	}

	// 生成 JWT token
	expirationTime := time.Now().Add(5 * time.Minute) // 设置 token 过期时间
	claims := &UserClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func main() {
	h := server.Default()

	renderHTML(h)

	h.POST("/register", register)

	h.POST("/login", login)

	h.POST("/display", display)

	h.Spin()
}
