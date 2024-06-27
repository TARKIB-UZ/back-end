package repo

// import (
// 	"context"

// 	"tarkib.uz/internal/entity"
// 	"tarkib.uz/pkg/postgres"
// )

// type ReceiptRepo struct {
// 	*postgres.Postgres
// }

// func NewReceiptRepo(pg *postgres.Postgres) *ReceiptRepo {
// 	return &ReceiptRepo{pg}
// }

// func (a *ReceiptRepo) Create(ctx context.Context, receipt *entity.Recipe) (*entity.User, error) {
// 	sql, args, err := a.Builder.
// 		Insert("recipes").
// 		Columns("id, first_name, last_name, phone_number, nickname, password, avatar").
// 		Values(user.ID, user.FirstName, user.LastName, user.PhoneNumber, user.NickName, user.Password, user.Avatar).
// 		ToSql()
// 	if err != nil {
// 		return nil, err
// 	}

// 	_, err = a.Pool.Exec(ctx, sql, args...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return user, nil
// }
