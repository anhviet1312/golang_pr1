package services

import (
	"context"
	"database/sql"
	"demo-cosebase/internal/datastore"
	"demo-cosebase/internal/datastore/redis_store"
	"demo-cosebase/internal/models"
	"demo-cosebase/pkg"
	"demo-cosebase/pkg/caching"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
)

type ServiceUser struct {
	container     *do.Injector
	redisDB       redis.UniversalClient
	postgresDB    *bun.DB
	readonlyCache caching.ReadOnlyCache
	cache         caching.Cache
}

func NewServiceUser(container *do.Injector) (*ServiceUser, error) {
	db, err := do.InvokeNamed[redis.UniversalClient](container, "redis-db")
	if err != nil {
		return nil, err
	}

	postgresDB, err := do.Invoke[*bun.DB](container)
	if err != nil {
		return nil, err
	}

	readonlyCache, err := do.Invoke[caching.ReadOnlyCache](container)
	if err != nil {
		return nil, err
	}

	cache, err := do.Invoke[caching.Cache](container)
	if err != nil {
		return nil, err
	}

	return &ServiceUser{container, db, postgresDB, readonlyCache, cache}, nil
}

func (service *ServiceUser) Authenticate(ctx context.Context, username string, password string) (*models.User, error) {
	user, err := datastore.FindUserByUsername(ctx, service.postgresDB, username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

func (service *ServiceUser) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	callback := func() (*models.User, error) {
		return datastore.FindUserByUsername(ctx, service.postgresDB, username)
	}
	return caching.UseCacheWithRO(ctx, service.readonlyCache, service.cache, DBKeyUserByUsername(username), CacheTtl5Mins, callback)
}

func (service *ServiceUser) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := datastore.FindUserByEmail(ctx, service.postgresDB, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (service *ServiceUser) CheckIfUserExists(ctx context.Context, email, username string) (*models.User, error) {
	user, err := datastore.FindUserByEmailOrUsername(ctx, service.postgresDB, email, username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return user, nil
}

func (service *ServiceUser) CreateUser(ctx context.Context, email, username, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUserId := pkg.GenerateRandomID()

	go func() {
		err := service.sendMailActiveCode(ctx, email, newUserId)
		if err != nil {
			fmt.Println(err)
		}
	}()

	user := &models.User{
		ID:       newUserId,
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		IsActive: false,
	}
	_, err = datastore.CreateUser(ctx, service.postgresDB, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (service *ServiceUser) sendMailActiveCode(ctx context.Context, email string, userID int64) error {
	otpCode, err := pkg.GenerateOTP()
	if err != nil {
		return err
	}

	emailUsername := os.Getenv("EMAIL_USERNAME")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	auth := sasl.NewPlainClient("", emailUsername, emailPassword)

	emailNotice := pkg.Email{
		From:    emailUsername,
		To:      []string{email},
		Subject: "Kích hoạt tài khoản",
		Body:    fmt.Sprintf("Mã kích hoạt tài khoản của bạn là: %s", otpCode),
	}
	err = pkg.SendMail(&emailNotice, auth)
	if err != nil {
		return err
	}

	_, err = redis_store.SetOtpCode(ctx, service.redisDB, userID, otpCode)
	return err
}

func (service *ServiceUser) ActivateUser(ctx context.Context, req *models.ActivationRequest) (bool, error) {
	user, err := datastore.FindUserByID(ctx, service.postgresDB, req.UserID)
	if err != nil {
		return false, err
	}

	if user.IsActive {
		return false, fmt.Errorf("user already activated")
	}

	otpCode, err := redis_store.GetOtpCode(ctx, service.redisDB, req.UserID)
	if err != nil {
		return false, err
	}

	if otpCode != req.ActivationCode {
		return false, fmt.Errorf("invalid otp code")
	}

	user.IsActive = true
	_, err = datastore.UpdateUser(ctx, service.postgresDB, user)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (service *ServiceUser) GetUserInfo(accessToken string) (map[string]interface{}, error) {
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

func (service *ServiceUser) FindOrCreateUserByEmail(ctx context.Context, userInfo map[string]interface{}) (*models.User, error) {
	email, ok := userInfo["email"]
	if !ok {
		return nil, errors.New("email not found")
	}

	userEmail, ok := email.(string)
	if !ok {
		return nil, errors.New("email is not in valid format")
	}

	user, err := datastore.FindUserByEmail(ctx, service.postgresDB, userEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if user == nil {
		newUser := &models.User{
			ID:        pkg.GenerateRandomID(),
			Username:  userEmail,
			Email:     userEmail,
			FirstName: userInfo["given_name"].(string),
			LastName:  userInfo["family_name"].(string),
			IsActive:  true,
		}
		_, err = datastore.CreateUser(ctx, service.postgresDB, newUser)
		if err != nil {
			return nil, err
		}
		return newUser, nil
	}
	return user, nil
}
