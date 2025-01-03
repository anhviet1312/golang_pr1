package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func authorize(container *do.Injector) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			return next(c)
		}
	}
}
