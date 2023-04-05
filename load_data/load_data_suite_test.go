package main_test

import (
	. "go_test/load_data"
	"os"
	"testing"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoadData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LoadData Suite")
}

var _ = Describe("LoadData", func() {

	Describe("創建檔案", func() {
		Context("成功創建", func() {
			BeforeEach(func() {
				Create_file()
			})
			It("則檔案存在", func() {
				_, err := os.Stat(FilePath)
				Expect(err).Should(BeNil())
			})
		})
	})

	Describe("爬資料", func() {
		Context("成功抓到資料", func() {
			var article, id string
			BeforeEach(func() {
				article, id = Crawl()
			})
			It("檔案應該存在", func() {
				Expect(article).ShouldNot(BeEmpty())
				Expect(id).ShouldNot(BeEmpty())
			})
		})
	})

	Describe("資料庫處理", func() {
		Context("成功建立資料庫及資料表", func() {
			BeforeEach(func() {
				CreateDb("test")
				CreateTable()
			})

			It("資料庫應該存在", func() {
				_, flag := DB.Exec("DB_ID(`test`)")
				Expect(flag).ShouldNot(BeNil())
			})

			It("資料表應該存在", func() {
				_, flag := DB.Exec("DB_OBJECT(`test`.`recommend`)")
				Expect(flag).ShouldNot(BeNil())
			})

		})

		Context("成功放入資料", func() {
			var id, next string
			BeforeEach(func() {
				First_load()
				row := DB.QueryRow("SELECT `id`,`next` FROM `user`.`recommend` ;")
				_ = row.Scan(&id, &next)
			})

			It("資料應該存在", func() {
				Expect(id).ShouldNot(BeNil())
			})

			It("next這欄應該有資料", func() {
				Expect(next).ShouldNot(BeNil())
			})

		})

		Context("成功刪除ttl為0的資料", func() {
			It("資料應該消失", func() {
				var id string
				Updatettl(1)
				row := DB.QueryRow("SELECT `id` FROM `user`.`recommend` where `ttl`=0;")
				err := row.Scan(&id)
				Expect(err).Should(Equal(sql.ErrNoRows))
			})
		})

		AfterEach(func() {
			DeleteDb("test")
		})
	})

})
