package models

import "tarkib.uz/internal/entity"

type LoginRequest struct {
	NickName    string `json:"nickname"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginUser struct {
	ID          string
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	User        LoginUser `json:"user"`
}

type RegisterUser struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	NickName    string `json:"nickname"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	Avatar      string `json:"avatar"`
}

type VerifyUser struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

type VerifyUserResponse struct {
	User *entity.User `json:"user"`
}

type ForgotPasswordRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type ResetPasswordRequest struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}
