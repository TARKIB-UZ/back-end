// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"tarkib.uz/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	Auth interface {
		Register(context.Context, *entity.User) error
		Verify(context.Context, entity.VerifyUser) (*entity.User, error)
		ForgotPassword(context.Context, string) error
		ResetPassword(context.Context, string, string, string) error
	}

	AuthRepo interface {
		Create(context.Context, *entity.User) (*entity.User, error)
		CheckUser(context.Context, string) (bool, error)
		UpdatePassword(context.Context, string, string) error
	}

	// TranslationWebAPI -.
	AuthWebAPI interface {
		SendSMS(context.Context, string, string) error
		SendSMSWithAndroid(context.Context, string, string) error
		// Translate(entity.Translation) (entity.Translation, error)
	}
)
