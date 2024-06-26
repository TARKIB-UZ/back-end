package models

type RegisterUser struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	NickName    string `json:"nickname"`
	Password    string `json:"password"`
	Avatar      string `json:"avatar"`
}

type VerifyUser struct {
	PhoneNumber string
	Code        string
}

type VerifyUserResponse struct {
	ID          string
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
	AccessToken string
}

type ForgotPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type ResetPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}
