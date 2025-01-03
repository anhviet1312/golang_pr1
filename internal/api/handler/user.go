package handler

import (
	"demo-cosebase/internal/models"
	"demo-cosebase/internal/services"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hiendaovinh/toolkit/pkg/errorx"
	"github.com/hiendaovinh/toolkit/pkg/httpx-echo"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"net/http"
	"os"
	"time"
)

type groupUser struct {
	container *do.Injector
}

func (gr *groupUser) Login(c echo.Context) error {
	ctx := c.Request().Context()
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed"})
	}

	serviceUser, err := do.Invoke[*services.ServiceUser](gr.container)
	if err != nil {
		return httpx.RestAbort(c, nil, errorx.Wrap(err, errorx.Service))
	}

	user, err := serviceUser.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		return err
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token valid for 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	return c.JSON(http.StatusOK, models.LoginResponse{Token: tokenString})
}
