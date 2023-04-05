# Key-value-list

本次實作的方式是利用colly套件爬取ptt棒球版的資料，以該文章網址為page的key(實作過程用id表示)以及該文章的標題為page的內容，建立“共用的 Key-Value 列表系統”
，主要分為兩個資料夾load_data及api，資料庫使用aws的RDS資料庫使用的是MySql，考慮到情境為使用量龐大及每小時更新，因此選用Sql資料庫而非NoSql

## Load_data：爬取資料及建立、更新RDS資料庫
1.	爬取資料並寫入csv檔，紀錄為{id,article}的格式
2.	寫入資料庫並存為{id, article, next, ttl}的格式，其中next存放下一篇文章的id，ttl則為該文章能夠存活的天數
3.	每次會重新爬取並更新資料庫，若文章存在則增加其ttl，否則插入該文章，並每次會將ttl減1，若ttl到0則刪除該文章

## Api：建立RESTful api並使用 aws的Lambda及Api Gateway
1.  GetHead：取得資料庫中首篇文章id  
2.  GetPage：利用QueryString取得對應id的資訊  
3.  將其部署在aws的Lambda中並利用Api Gateway作為入口，讓其能實際運作，在網址後面加上/GetHead或/GetPage?input=”*要查詢的id*”即可使用  
4.  網址：https://b4743ws0hf.execute-api.ap-northeast-1.amazonaws.com/rest-api01  

## Unit Test：使用ginko框架為大部分函數進行測試

## 可改進方向
1.  目前仍需手動更新資料庫，可結合S3及Lambda達成自動更新
2.  改成使用grpc實作
