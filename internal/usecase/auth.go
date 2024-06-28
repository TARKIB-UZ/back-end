package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"tarkib.uz/config"
	"tarkib.uz/internal/entity"
	avatargenerator "tarkib.uz/pkg/avatar-generator"
	avatar "tarkib.uz/pkg/base64-image"
	"tarkib.uz/pkg/password"
	tokens "tarkib.uz/pkg/token"
)

type AuthUseCase struct {
	repo        AuthRepo
	webAPI      AuthWebAPI
	cfg         *config.Config
	RedisClient *redis.Client
	MinioClient *minio.Client
}

func NewAuthUseCase(r AuthRepo, w AuthWebAPI, cfg *config.Config, RedisClient *redis.Client, minioClient *minio.Client) *AuthUseCase {
	return &AuthUseCase{
		repo:        r,
		webAPI:      w,
		cfg:         cfg,
		RedisClient: RedisClient,
		MinioClient: minioClient,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, user *entity.User) error {
	var (
		userForRedis entity.UserForRedis
		imageName    string
	)
	IsExist, err := uc.repo.CheckField(ctx, "nickname", user.NickName)
	if err != nil {
		return err
	}

	if IsExist {
		return errors.New("this nickname is already taken")
	}

	IsExist, err = uc.repo.CheckField(ctx, "phone_number", user.PhoneNumber)
	if err != nil {
		return err
	}

	if IsExist {
		return errors.New("user with this phone number already registered")
	}

	if user.Avatar == "" {
		imageName = uuid.NewString() + ".png"
		initials := avatargenerator.GetInitial(user.FirstName, user.LastName)
		avatargenerator.CreateProfileImage(initials, imageName)
		file, err := os.Open(imageName)
		if err != nil {
			return err
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", imageName)
		if err != nil {
			fmt.Println("Error writing to form:", err)
			return err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			fmt.Println("Error copying file:", err)
			return err
		}

		err = writer.Close()
		if err != nil {
			fmt.Println("Error closing writer:", err)
			return err
		}

		req, err := http.NewRequest("POST", os.Getenv("FILE_UPLOAD_URL"), body)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return err
		}
		defer resp.Body.Close()

		var result struct {
			Url string `json:"url"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			return err
		}

		fmt.Println("Response:", result)
		userForRedis.Avatar = result.Url
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		temp := r.Intn(1000000)

		code := fmt.Sprintf("%06d", temp)

		userForRedis.ID = uuid.NewString()
		userForRedis.FirstName = user.FirstName
		userForRedis.LastName = user.LastName
		userForRedis.NickName = user.NickName
		userForRedis.Password = user.Password
		userForRedis.PhoneNumber = user.PhoneNumber
		userForRedis.Code = code

		byteData, err := json.Marshal(userForRedis)
		if err != nil {
			return err
		}

		status := uc.RedisClient.Set(ctx, user.PhoneNumber, byteData, 10*time.Minute)
		if status.Err() != nil {
			return err
		}

		if err := uc.webAPI.SendSMSWithAndroid(ctx, user.PhoneNumber, code, "register"); err != nil {
			return err
		}

		if err := os.Remove(imageName); err != nil {
			return err
		}

		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	temp := r.Intn(1000000)

	code := fmt.Sprintf("%06d", temp)

	userForRedis.ID = uuid.NewString()
	userForRedis.Avatar = user.Avatar
	userForRedis.FirstName = user.FirstName
	userForRedis.LastName = user.LastName
	userForRedis.NickName = user.NickName
	userForRedis.Password = user.Password
	userForRedis.PhoneNumber = user.PhoneNumber
	userForRedis.Code = code

	byteData, err := json.Marshal(userForRedis)
	if err != nil {
		return err
	}

	if err := uc.webAPI.SendSMSWithAndroid(ctx, user.PhoneNumber, code, "register"); err != nil {
		return err
	}

	status := uc.RedisClient.Set(ctx, user.PhoneNumber, byteData, 10*time.Minute)
	if status.Err() != nil {
		return err
	}

	return nil

}

func (uc *AuthUseCase) Verify(ctx context.Context, request entity.VerifyUser) (*entity.User, error) {
	var (
		userForRedis entity.UserForRedis
	)
	endpoint := os.Getenv("SERVER_IP")
	data := uc.RedisClient.Get(ctx, request.PhoneNumber)
	if data.Err() != nil {
		return nil, data.Err()
	}

	if data.Val() == "" {
		return nil, errors.New("verification code expired")
	}

	err := json.Unmarshal([]byte(data.Val()), &userForRedis)
	if err != nil {
		return nil, err
	}

	if userForRedis.Code != request.Code {
		return nil, errors.New("invalid verification code")
	}

	avatarImage := uuid.NewString() + ".png"
	if err := avatar.SaveAvatar(userForRedis.Avatar, avatarImage, uc.MinioClient); err != nil {
		return nil, err
	}

	jwtHandler := tokens.JWTHandler{
		Sub:       userForRedis.ID,
		Iss:       time.Now().UTC().Format(time.RFC3339),
		Exp:       time.Now().UTC().Add(time.Hour * 168).Format(time.RFC3339),
		Role:      "user",
		SigninKey: uc.cfg.Casbin.SigningKey,
		Timeout:   uc.cfg.Casbin.AccessTokenTimeOut,
	}

	access, _, err := jwtHandler.GenerateAuthJWT()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := password.HashPassword(userForRedis.Password)
	if err != nil {
		return nil, err
	}

	_, err = uc.repo.Create(ctx, &entity.User{
		ID:          userForRedis.ID,
		FirstName:   userForRedis.FirstName,
		LastName:    userForRedis.LastName,
		PhoneNumber: userForRedis.PhoneNumber,
		NickName:    userForRedis.NickName,
		Password:    hashedPassword,
		Avatar:      fmt.Sprintf("https://%s/%s/%s", endpoint, "avatars", avatarImage),
		AccessToken: access,
	})
	if err != nil {
		return nil, err
	}

	return &entity.User{
		ID:          userForRedis.ID,
		FirstName:   userForRedis.FirstName,
		LastName:    userForRedis.LastName,
		PhoneNumber: userForRedis.PhoneNumber,
		NickName:    userForRedis.NickName,
		Password:    userForRedis.Password,
		Avatar:      userForRedis.Avatar,
		AccessToken: access,
	}, nil
}

func (uc *AuthUseCase) ForgotPassword(ctx context.Context, phoneNumber string) error {
	IsExists, err := uc.repo.CheckField(ctx, "phone_number", phoneNumber)
	if err != nil {
		return err
	}

	if IsExists {
		return errors.New("this phone number not registered in tarkib.uz yet")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	temp := r.Intn(1000000)
	code := fmt.Sprintf("%06d", temp)

	status := uc.RedisClient.Set(ctx, phoneNumber+"_reset", code, 10*time.Minute)
	if status.Err() != nil {
		return status.Err()
	}

	if err := uc.webAPI.SendSMSWithAndroid(ctx, phoneNumber, code, "forgot"); err != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) ResetPassword(ctx context.Context, phoneNumber, code, newPassword string) error {
	storedCode := uc.RedisClient.Get(ctx, phoneNumber+"_reset")
	if storedCode.Err() != nil {
		return storedCode.Err()
	}

	if storedCode.Val() != code {
		return errors.New("invalid reset code")
	}

	hashedPassword, err := password.HashPassword(newPassword)
	if err != nil {
		return err
	}

	err = uc.repo.UpdatePassword(ctx, phoneNumber, hashedPassword)
	if err != nil {
		return err
	}

	uc.RedisClient.Del(ctx, phoneNumber+"_reset")

	return nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req entity.LoginRequest) (*entity.LoginResponse, error) {
	var user *entity.User
	var err error

	if req.NickName != "" {
		user, err = uc.repo.GetUserByNickName(ctx, req.NickName)
		if err != nil {
			return nil, errors.New("user not found")
		}
	} else {
		user, err = uc.repo.GetUserByPhoneNumber(ctx, req.PhoneNumber)
		if err != nil {
			return nil, errors.New("user not found")
		}
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if !password.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid password")
	}

	expDuration := time.Duration(uc.cfg.Casbin.AccessTokenTimeOut) * time.Second
	expTime := time.Now().Add(expDuration)

	jwtHandler := tokens.JWTHandler{
		Sub:       user.ID,
		Iss:       time.Now().String(),
		Exp:       expTime.String(),
		Role:      "user",
		SigninKey: uc.cfg.Casbin.SigningKey,
		Timeout:   uc.cfg.Casbin.AccessTokenTimeOut,
	}

	accessToken, _, err := jwtHandler.GenerateAuthJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	return &entity.LoginResponse{
		AccessToken: accessToken,
		User: entity.LoginUser{
			ID:          user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			PhoneNumber: user.PhoneNumber,
			NickName:    user.NickName,
			Password:    user.Password,
			Avatar:      user.Avatar,
		},
	}, nil
}
