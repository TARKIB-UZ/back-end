package entity

type User struct {
	ID          string
	FirstName   string
	LastName    string
	PhoneNumber string
	NickName    string
	Password    string
	Avatar      string
	AccessToken      string
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
