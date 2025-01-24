package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/utkarsh867/cowg/cowg"
)

func RunHTTPServer(co *cowg.Cowg) {
  e := echo.New()

  e.GET("/config", func(c echo.Context) error {
    fileName := "wg.conf"
    c.Response().Header().Set("Content-Disposition", "attachment; filename="+fileName)
    c.Response().Header().Set("Content-Type", "text/plain")
    c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(co.Config)))
    return c.String(http.StatusOK, co.Config)
  })

  e.Logger.Fatal(e.Start(":22080"))
}
