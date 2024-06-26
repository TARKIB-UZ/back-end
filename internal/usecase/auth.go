package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"tarkib.uz/config"
	"tarkib.uz/internal/entity"
	"tarkib.uz/pkg/password"
	tokens "tarkib.uz/pkg/token"
)

type AuthUseCase struct {
	repo        AuthRepo
	webAPI      AuthWebAPI
	cfg         *config.Config
	RedisClient *redis.Client
}

func NewAuthUseCase(r AuthRepo, w AuthWebAPI, cfg *config.Config, RedisClient *redis.Client) *AuthUseCase {
	return &AuthUseCase{
		repo:        r,
		webAPI:      w,
		cfg:         cfg,
		RedisClient: RedisClient,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, user *entity.User) error {
	var userForRedis entity.UserForRedis
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

	jwtHandler := tokens.JWTHandler{
		Sub:       userForRedis.ID,
		Iss:       time.Now().String(),
		Exp:       time.Now().Add(time.Hour * 168).String(),
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
		Avatar:      userForRedis.Avatar,
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
