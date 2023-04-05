package main_test

import (
	"encoding/json"
	"fmt"
	. "go_test/api"
	"io"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

var _ = Describe("Api", func() {
	// BeforeEach(func() {
	// 	Main()
	// })
	Describe("GetHead", func() {
		Context("成功得到head", func() {
			res, err := http.Get("http://127.0.0.1/GetHead")
			if err != nil {
				fmt.Println("Error:", err)
			}
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)
			It("則能成功get", func() {
				Expect(err).Should(BeNil())
			})
			It("則能得到首個文章", func() {
				first := Find_first()
				Expect(string(b)).Should(Equal(first))
			})
		})
	})

	Describe("GetPage", func() {
		Context("成功得到Page", func() {
			var article Article
			first := Find_first()
			url := fmt.Sprintf("http://127.0.0.1/GetPage?input=%s", first)
			res, err := http.Get(url)
			if err != nil {
				fmt.Println("Error:", err)
			}
			defer res.Body.Close()
			json.NewDecoder(res.Body).Decode(&article)
			It("則能成功get", func() {
				Expect(err).Should(BeNil())
			})
			It("則能得到指定文章", func() {
				a := Load(first)
				Expect(article.Id).Should(Equal(a.Id))
				Expect(article.Content).Should(Equal(a.Content))
				Expect(article.Next).Should(Equal(a.Next))
			})
		})
	})
})
