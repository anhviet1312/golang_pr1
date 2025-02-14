package handler

import (
	"context"
	"demo-cosebase/internal/models"
	"demo-cosebase/internal/services"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hiendaovinh/toolkit/pkg/errorx"
	"github.com/hiendaovinh/toolkit/pkg/httpx-echo"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
	"time"
)

const (
	ExpireTokenDuration = time.Minute * 2
)

var googleOauthConfig = &oauth2.Config{}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"profile", "email"}, // Adjust scopes as needed
		Endpoint:     google.Endpoint,
	}
}

type groupUser struct {
	container *do.Injector
}

func (gr *groupUser) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.RegisterRequest
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

	existingUser, err := serviceUser.CheckIfUserExists(ctx, req.Email, req.Username)
	if err != nil {
		return httpx.RestAbort(c, nil, errorx.Wrap(err, errorx.Database))
	}
	if existingUser != nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Email or username already exists"})
	}

	// Register the user
	newUser, err := serviceUser.CreateUser(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		return httpx.RestAbort(c, nil, err)
	}

	return c.JSON(http.StatusOK, models.RegisterResponse{Message: "Registration successful", User: newUser})
}

func (gr *groupUser) ActivateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.ActivationRequest
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

	_, err = serviceUser.ActivateUser(ctx, &req)
	if err != nil {
		return httpx.RestAbort(c, nil, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Account activated successfully"})
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
		"id":         fmt.Sprintf("%d", user.ID),
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"exp":        time.Now().Add(ExpireTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, models.LoginResponse{Token: tokenString})
}

func (gr *groupUser) GoogleCallbackHandlerLogin(c echo.Context) error {
	code := c.QueryParam("code")
	googleToken, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return httpx.RestAbort(c, nil, errorx.Wrap(fmt.Errorf("exchage code to googleToken fail"), errorx.Invalid))
	}

	servicesUser, err := do.Invoke[*services.ServiceUser](gr.container)
	if err != nil {
		return httpx.RestAbort(c, nil, errorx.Wrap(err, errorx.Service))
	}

	userInfo, err := servicesUser.GetUserInfo(googleToken.AccessToken)
	if err != nil {
		return httpx.RestAbort(c, nil, err)
	}

	user, err := servicesUser.FindOrCreateUserByEmail(c.Request().Context(), userInfo)
	if err != nil {
		return httpx.RestAbort(c, nil, errorx.Wrap(err, errorx.Service))
	}

	claims := jwt.MapClaims{
		"id":         fmt.Sprintf("%d", user.ID),
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"exp":        time.Now().Add(ExpireTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, tokenString)
}
