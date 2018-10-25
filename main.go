package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/:jan", func(context echo.Context) error {
		jan := context.Param("jan")
		if jan == ""{
			return context.JSON(http.StatusBadRequest,"Bad")
		}
		// client
		client := &http.Client{}
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		// Request
		req, err := http.NewRequest("GET", "https://mnrate.com/search?i=All&kwd="+ jan +"&s=", nil)
		if err != nil {
			panic(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		r := struct {
			URL string `json:"url"`
			ASIN string `json:"asin"`
		}{}
		u := resp.Header.Get("Location")
		r.URL = u
		r.ASIN = u[len(u)-10:]

		return context.JSON(http.StatusOK,r)
	})

	e.Logger.Fatal(e.Start(":8080"))
}