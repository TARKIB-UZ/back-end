package entity

type User struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	NickName    string `json:"nickname"`
	Password    string `json:"password"`
	Avatar      string `json:"avatar"`
	AccessToken string `json:"access_token"`
}

type UserForRedis struct {
	ID          string
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
	Code        string
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

type LoginUser struct {
	ID          string
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
}

type LoginRequest struct {
	NickName    string `json:"nickname"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	User        LoginUser `json:"user"`
}
