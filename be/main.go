package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "root")
	})
	e.Logger.Fatal(e.Start(":" + os.Getenv("GOLANG_PORT")))
}
