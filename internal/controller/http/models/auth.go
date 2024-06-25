package models

type RegisterUser struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
}

type VerifyUser struct {
	PhoneNumber string
	Code        string
}

type VerifyUserResponse struct {
	ID string
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
	AccessToken string
}