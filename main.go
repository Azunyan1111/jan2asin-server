package main

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"net/url"
)

var goToGetKey bool
var saveKey string

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	goToGetKey = true

	e.GET("/:jan", func(context echo.Context) error {
		jan := context.Param("jan")
		if jan == ""{
			return context.JSON(http.StatusBadRequest,"Bad:" + jan)
		}
		// client
		var key string
		var err error
		if goToGetKey{
			key,err = getKey()
			if err != nil{
				return context.JSON(http.StatusInternalServerError,err.Error())
			}
			saveKey = key
			goToGetKey = true
		}else{
			key = saveKey
		}
		asin,err := janToAsin(jan,key)
		if err != nil{
			return context.JSON(http.StatusInternalServerError,err.Error())
		}
		r := struct {
			ASIN string `json:"asin"`
		}{}
		r.ASIN = asin

		return context.JSON(http.StatusOK,r)
	})

	e.Logger.Fatal(e.Start(":8080"))
}


func janToAsin(jan string,key string)(string,error){
	//https://sellercentral.amazon.co.jp/fba/profitabilitycalculator/productmatches?searchKey=4549526605444&language=ja_JP&profitcalcToken=gj8%2B3vyvJX1%2Fmwj8jTfkpgFEeZs3M4JMrMbp79QAAAAJAAAAAFvfA1tyYXcAAAAAFVfwLBerPie4v1Ep%2F%2F%2F%2F
	u := url.Values{}
	u.Set("searchKey",jan)
	u.Add("language","ja_JP")
	u.Add("profitcalcToken",key)
	req, _ := http.NewRequest("GET", "https://sellercentral.amazon.co.jp/fba/profitabilitycalculator/productmatches?" + u.Encode(), nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36")
	req.Header.Set("accept-language", "ja,en;q=0.9")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil{
		return "",errors.New("API使いすぎ" + err.Error())
	}

	var apiJson ApiJson

	if err := json.NewDecoder(resp.Body).Decode(&apiJson); err != nil{
		return "",errors.New("正しいJSON返って無いでしょ。てかパース無理だった。" + err.Error())
	}
	return apiJson.Data[0].Asin,nil
}

func getKey()(string,error){
	// ここにURLにアクセスしてCSRF鍵を入手
	u :=  "http://sellercentral.amazon.co.jp/hz/fba/profitabilitycalculator/index?lang=ja_JP"

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36")
	req.Header.Set("accept-language", "ja,en;q=0.9")

	client := new(http.Client)
	resp, err := client.Do(req)

	doc,err := goquery.NewDocumentFromResponse(resp)
	if err != nil{
		return "",errors.New("HTMLのパース失敗@CSRF鍵鳥")
	}
	key := ""
	doc.Find("input").Each(func(i int, selection *goquery.Selection) {
		name,e := selection.Attr("name")
		if !e{
			return
		}
		if name == "profitcalcToken"{
			val,e := selection.Attr("value")
			if !e{
				return
			}
			key = val
		}
	})
	if key == ""{
		return "",errors.New("HTMLの中に鍵が無い")
	}
	return key,nil
}

type ApiJson struct {
	Data []struct {
		Asin                   string  `json:"asin"`
	} `json:"data"`
}