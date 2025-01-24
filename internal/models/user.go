package models

import "github.com/uptrace/bun"

type User struct {
	bun.BaseModel `bun:"table:user"`
	ID            int64  `bun:"id,pk" json:"id"`
	FirstName     string `bun:"first_name" json:"first_name"`
	LastName      string `bun:"last_name" json:"last_name"`
	Username      string `bun:"username" json:"username"`
	Password      string `bun:"password" json:"password"`
	Email         string `bun:"email" json:"email"`
	IsActive      bool   `bun:"is_active" json:"is_active"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
	Username string `json:"username" form:"username" validate:"required"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	User    *User  `json:"user"`
}

type ActivationRequest struct {
	UserID         int64  `json:"user_id" validate:"required"`
	ActivationCode string `json:"activation_code" validate:"required"`
}
