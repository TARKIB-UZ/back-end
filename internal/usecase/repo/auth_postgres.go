package repo

import (
	"context"

	"github.com/Masterminds/squirrel"
	"tarkib.uz/internal/entity"
	"tarkib.uz/pkg/postgres"
)

type AuthRepo struct {
	*postgres.Postgres
}

func NewAuthRepo(pg *postgres.Postgres) *AuthRepo {
	return &AuthRepo{pg}
}

func (a *AuthRepo) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	sql, args, err := a.Builder.
		Insert("users").
		Columns("id, first_name, last_name, phone_number, nickname, password, avatar").
		Values(user.ID, user.FirstName, user.LastName, user.PhoneNumber, user.NickName, user.Password, user.Avatar).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = a.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *AuthRepo) CheckUser(ctx context.Context, nickname string) (bool, error) {
	var count int

	sql, args, err := a.Builder.
		Select("count(nickname)").
		From("users").
		Where(squirrel.Eq{
			"nickname": nickname,
		}).ToSql()
	if err != nil {
		return false, err
	}

	err = a.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (a *AuthRepo) UpdatePassword(ctx context.Context, phoneNumber, newPassword string) error {
	sql, args, err := a.Builder.
		Update("users").
		Set("password", newPassword).
		Where(squirrel.Eq{
			"phone_number": phoneNumber,
		}).ToSql()
	if err != nil {
		return err
	}

	_, err = a.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
