package handler

import (
	"github.com/hiendaovinh/toolkit/pkg/httpx-echo"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
	"net/http"
)

type Config struct {
	Container *do.Injector
	Mode      string
	Origins   []string
}

func New(cfg *Config) (http.Handler, error) {
	r := echo.New()
	r.Pre(middleware.RemoveTrailingSlash())
	if cfg.Mode == "debug" {
		r.Debug = true
		pprof.Register(r)
	}

	r.JSONSerializer = httpx.SegmentJSONSerializer{}
	r.Use(middleware.Recover())

	routesAPIv1 := r.Group("/api/v1")
	{
		cors := middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.Origins,
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
			AllowCredentials: true,
			MaxAge:           60 * 60,
		})
		routesAPIv1.Use(cors)
		routesAPIv1.GET("", Hello)
	}

	r.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "👻️")
	})
	return r, nil
}

func Hello(c echo.Context) error {
	return httpx.RestAbort(c, "hello world", nil)
}
