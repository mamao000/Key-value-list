package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
)

var Pwd string
var FilePath string
var FileName = "data.csv"

var DB *sql.DB

// 連接資料庫
func init() {
	Pwd, _ = os.Getwd()
	FilePath = filepath.Join(Pwd, FileName)

	driver := "mysql"
	user := "admin"
	password := "hk4g4001"
	endpoint := "test-db1.c6sot98e4zta.ap-northeast-1.rds.amazonaws.com"
	port := "3306"
	dbName := "mydb"
	charset := "charset=utf8mb4"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, password, endpoint, port, dbName, charset)
	//fmt.Println(dsn)
	dbConnect, err := sql.Open(driver, dsn)

	if err != nil {
		log.Fatalln(err)
	}

	err = dbConnect.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	DB = dbConnect

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(10)
}

func main() {
	_, err := os.Stat(FilePath)
	if os.IsNotExist(err) {
		Create_file()
	}
	CreateDb("`user`")
	CreateTable()
	//DeleteDb("`user`")
	First_load()
}

// 爬資料並寫入csv檔
func Crawl() (string, string) {
	var article, id string
	c := colly.NewCollector()

	f, err := os.OpenFile(FilePath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")

	c.OnHTML(".title", func(e *colly.HTMLElement) {
		article, id = e.Text, e.ChildAttr("a", "href")
		if id == "" {
			return
		}
		id = id[14 : len(id)-5]
		Write(article, id)
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36")
	})

	c.Visit("https://www.ptt.cc/bbs/Baseball/index.html")
	return article, id
}

// 建檔
func Create_file() {
	f, err := os.Create(FilePath)
	defer f.Close()
	if err != nil {
		log.Fatalln("create file error: ", err)
	}
	f.WriteString("\xEF\xBB\xBF")
}

// 寫入資料到csv
func Write(article, id string) {
	file, err := os.OpenFile(FilePath, os.O_APPEND, 0777)
	if err != nil {
		log.Fatalln("找不到CSV檔案路徑:", FilePath, err)
	}

	w := csv.NewWriter(file)
	x := []string{id, article}
	w.Write(x)

	w.Flush()
}

// 建資料庫
func CreateDb(dbName string) {
	_, err := DB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print(err)
}

// 刪除資料庫
func DeleteDb(dbName string) {
	_, err := DB.Exec("DROP DATABASE IF EXISTS " + dbName + ";")
	if err != nil {
		log.Fatalln(err)
	}
}

// 建table
func CreateTable() {
	_, err := DB.Exec("CREATE TABLE IF NOT EXISTS `user`.`recommend`(`id` VARCHAR(100) PRIMARY KEY NOT NULL,`article` VARCHAR (200),`next` VARCHAR(100),`ttl` INT(2))")
	if err != nil {
		log.Fatalln(err)
	}
}

// 第一次載入
func First_load() {
	Update(0)
}

// 每日更新
func Daily_update() {
	Update(1)
}

// 初次載入資料進table
// func load() {
// 	file, err := os.OpenFile(FilePath, os.O_RDONLY, 0777)
// 	if err != nil {
// 		log.Fatalln("找不到CSV檔案路徑:", FilePath, err)
// 	}

// 	r := csv.NewReader(file)
// 	r.Comma = ','
// 	var id, article string
// 	for {
// 		record, err := r.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		id = record[0]
// 		article = record[1]
// 		_, err = db.Exec("insert INTO `user`.`recommend`(id,article,next,ttl) values(?,?,?,?);", id, article, nil, 1)
// 		if err != nil {
// 			fmt.Printf("Insert data failed,err:%v", err)
// 			return
// 		}
// 	}
// 	set_next()
// }

// 更新資料庫存在的文章更新ttl不存在則insert
func Update(times int) {
	Crawl()                                               //先爬資料
	file, err := os.OpenFile(FilePath, os.O_RDONLY, 0777) //讀檔
	if err != nil {
		log.Fatalln("找不到CSV檔案路徑:", FilePath, err)
	}

	r := csv.NewReader(file)
	r.Comma = ','
	var id, article string
	for {
		record, err := r.Read() //讀檔到EOF
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		id = record[0]
		article = record[1]
		row := DB.QueryRow("SELECT `id` FROM `user`.`recommend` WHERE `id` = ?;", record[0]) //尋找是否存在
		err = row.Scan(&id)
		if err == sql.ErrNoRows { //沒有則insert
			_, err := DB.Exec("insert INTO `user`.`recommend`(id,article,next,ttl) values(?,?,?,?)", id, article, nil, 1)
			if err != nil {
				fmt.Printf("Insert data failed,err:%v", err)
				return
			}
		} else if err != nil {
			log.Fatalln(err)
		} else { //有則更新ttl加1
			Addttl(id, 1)
		}
	}
	Updatettl(times)
	Set_next()
}

//根據id刪除
// func delete_item(s string) {
// 	_, err := db.Exec("DELETE FROM `user`.`recommend` WHERE `id` = ?;", s)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// }

// 根據ttl刪除
func Delete_ttl() {
	_, err := DB.Exec("DELETE FROM `user`.`recommend` WHERE `ttl` = 0;")
	if err != nil {
		log.Fatalln(err)
	}
}

// 指定id其ttl加times
func Addttl(s string, times int) {
	_, err := DB.Exec("UPDATE `user`.`recommend` SET `ttl`= `ttl`+? WHERE `id` = ?;", times, s)
	if err != nil {
		log.Fatalln(err)
	}
}

// 更新所有ttl減times 並刪除ttl為0的資料
func Updatettl(times int) {
	_, err := DB.Exec("UPDATE `user`.`recommend` SET `ttl`= `ttl`-? ;", times)
	if err != nil {
		log.Fatalln(err)
	}
	Delete_ttl()
}

// 設定所有next值
func Set_next() {
	count := 0
	var past, cur string
	rows, err := DB.Query("SELECT `id` FROM `user`.`recommend`;")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		if count == 0 {
			err = rows.Scan(&past)
			if err != nil {
				log.Fatalln(err)
			}
			count++
			continue
		}
		err = rows.Scan(&cur)
		if err != nil {
			log.Fatalln(err)
		}
		_, err := DB.Exec("UPDATE `user`.`recommend` SET `next`= ? WHERE `id`=?;", cur, past)
		if err != nil {
			log.Fatalln(err)
		}
		past = cur
	}
	defer rows.Close()
}
