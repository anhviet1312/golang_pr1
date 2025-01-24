package handler

import (
	"github.com/go-playground/validator/v10"
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

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func New(cfg *Config) (http.Handler, error) {
	r := echo.New()
	r.Validator = &CustomValidator{validator: validator.New()}
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
		routesAPIv1.GET("", Hello, JWTMiddleware(cfg.Container))

		routesAPIv1User := routesAPIv1.Group("/user")
		{
			u := groupUser{cfg.Container}
			//routesAPIv1User.Use(JWTMiddleware(cfg.Container))
			routesAPIv1User.GET("/me", Hello)
			routesAPIv1User.POST("/login", u.Login)
			routesAPIv1User.POST("/register", u.Register)
			routesAPIv1User.POST("/activate", u.ActivateUser)

		}

	}

	r.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "👻️")
	})
	return r, nil
}

func Hello(c echo.Context) error {
	//Retrieve the user_id from the context
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User ID not found in token"})
	}

	// Respond with a message including the user_id
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Hello, user!",
		"user_id": userID,
	})
}
