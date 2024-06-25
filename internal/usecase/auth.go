package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"tarkib.uz/config"
	"tarkib.uz/internal/entity"
	tokens "tarkib.uz/pkg/token"
)

// AuthUseCase -.
type AuthUseCase struct {
	repo        AuthRepo
	webAPI      AuthWebAPI
	cfg         *config.Config
	RedisClient *redis.Client
}

// New -.
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
	IsExist, err := uc.repo.CheckUser(ctx, user.NickName)
	if err != nil {
		return err
	}

	if IsExist {
		return errors.New("user already exists")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	temp := r.Intn(1000000)

	code := fmt.Sprintf("%06d", temp)

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

	if err := uc.webAPI.SendSMS(ctx, user.PhoneNumber, code); err != nil {
		return err
	}

	status := uc.RedisClient.Set(ctx, user.PhoneNumber, byteData, 10*time.Minute)
	if status.Err() != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) Verify(ctx context.Context, request entity.VerifyUser) (*entity.VerifyUserResponse, error) {
	var (
		userForRedis  entity.UserForRedis
		userWithToken entity.VerifyUserResponse
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

	userWithToken.AccessToken = access
	userWithToken.FirstName = userForRedis.FirstName
	userWithToken.LastName = userForRedis.LastName
	userWithToken.NickName = userForRedis.NickName
	userWithToken.Password = userForRedis.Password
	userWithToken.PhoneNumber = userForRedis.PhoneNumber
	userWithToken.Avatar = userForRedis.Avatar

	return &userWithToken, nil
}
