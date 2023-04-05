package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var ginLambda *ginadapter.GinLambda

type Article struct {
	Id      string
	Content string
	Next    string
}

// 連接rds資料庫
func init() {
	driver := "mysql"
	user := "admin"
	password := "hk4g4001"
	endpoint := "test-db1.c6sot98e4zta.ap-northeast-1.rds.amazonaws.com"
	port := "3306"
	dbName := "mydb"
	charset := "charset=utf8mb4"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, password, endpoint, port, dbName, charset)
	dbConnect, err := sql.Open(driver, dsn)

	if err != nil {
		log.Fatalln(err)
	}

	err = dbConnect.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	db = dbConnect

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, request)
}

func main() {
	g := gin.Default()
	g.GET("/GetPage", Getpage)
	g.GET("/GetHead", Gethead)
	ginLambda = ginadapter.New(g)
	lambda.Start(Handler)
}

// getpage api
func Getpage(c *gin.Context) {
	input := c.Query("input")
	a := Load(input)
	if a.Id == "" {
		str := fmt.Sprintf("%s doesn't exist", input)
		msg := []byte(str)
		c.Data(http.StatusOK, "text/plain", msg)
	} else {
		j, _ := json.Marshal(a)
		c.Data(http.StatusOK, "application/json", j)
	}
}

// gethead api
func Gethead(c *gin.Context) {
	id := Find_first()
	msg := []byte(id)
	c.Data(http.StatusOK, "text/plain", msg)
}

// load 指定id值的article及next
func Load(s string) Article {
	rows, err := db.Query("SELECT `id`, `article`,`next` FROM `user`.`recommend` WHERE `id` = ?;", s)
	defer rows.Close()
	var id, article string
	var next sql.NullString
	for rows.Next() {
		err = rows.Scan(&id, &article, &next)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return Article{Id: id, Content: article, Next: next.String}
}

// 找到首項
func Find_first() string {
	rows, err := db.Query("SELECT `id` FROM `user`.`recommend` ORDER BY `id` LIMIT 1;")
	var id string
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer rows.Close()
	return id
}
