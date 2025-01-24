package handler

import (
	"context"
	"demo-cosebase/internal/models"
	"demo-cosebase/internal/services"
	"encoding/json"
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
	//fmt.Println(user)
	claims := jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", user.ID),
		"exp":     time.Now().Add(time.Minute * 5).Unix(),
	}
	fmt.Println(claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	fmt.Println(tokenString)
	return c.JSON(http.StatusOK, models.LoginResponse{Token: tokenString})
}

func GoogleCallbackHandlerLogin(c echo.Context) error {
	code := c.QueryParam("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)

	if err != nil {
		return httpx.RestAbort(c, nil, errorx.Wrap(fmt.Errorf("exchage code to token fail"), errorx.Invalid))
	}

	userInfo, err := GetUserInfo(token.AccessToken)
	if err != nil {
		return httpx.RestAbort(c, nil, err)
	}

	userEmail, ok := userInfo["email"]
	if !ok {
		return httpx.RestAbort(c, nil, errorx.Wrap(fmt.Errorf("user email not found"), errorx.Invalid))
	}

	userSevice := do.Invoke[*services.ServiceUser]

}

func GetUserInfo(accessToken string) (map[string]interface{}, error) {
	userInfoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s", userInfoEndpoint, accessToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
