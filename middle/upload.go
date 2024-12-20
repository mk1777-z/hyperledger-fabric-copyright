package middle

import (
	"context"
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

func Upload(_ context.Context, c *app.RequestContext) {
	var uploadInfo conf.Upload
	c.Bind(&uploadInfo)
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", conf.Con.Mysql.DbUser, conf.Con.Mysql.DbPassword, conf.Con.Mysql.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Database connection error"})
		log.Fatal("OPEN SQL ERROR")
		return
	}
	defer db.Close() // 确保数据库连接在结束时关闭

	rows, _ := db.Query("SELECT * FROM item WHERE name = ?", uploadInfo.Name)
	if rows != nil {
		c.Status(http.StatusConflict)
		c.JSON(http.StatusConflict, utils.H{"message": "Item Already Exist"})
	}

	_, err = db.Exec("INSERT INTO item (name,simple_dsc,dsc,price,img,start_time) VALUES (%s,%s,%s,%s,%s,%s)", uploadInfo.Name, uploadInfo.Simple_dsc, uploadInfo.Dsc, uploadInfo.Price, uploadInfo.Img, time.Now())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, utils.H{"message": "Internal Server Error"})
		log.Fatal("INSERT DATA ERROR")
		return
	}
	c.Status(http.StatusOK)
	c.JSON(http.StatusOK, utils.H{"message": "Create item success"})
}
