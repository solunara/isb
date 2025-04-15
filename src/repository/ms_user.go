package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/repository/dao"
)

type MsUserRepository interface {
	Create(ctx context.Context, u MsUser) error
	FindByUsername(ctx context.Context, username string) (MsUser, error)
}

type CachedMsUserRepository struct {
	dao dao.MsUserDAO
}

func NewMsUserRepository(dao dao.MsUserDAO) MsUserRepository {
	return &CachedMsUserRepository{
		dao: dao,
	}
}

func (repo *CachedMsUserRepository) Create(ctx context.Context, u MsUser) error {
	return repo.dao.Insert(ctx, repo.toModel(u))
}

func (repo *CachedMsUserRepository) FindByUsername(ctx context.Context, username string) (MsUser, error) {
	modelMsUser, err := repo.dao.FindByUsername(ctx, username)
	fmt.Println(err)
	if err != nil {
		return MsUser{}, err
	}
	return repo.toView(modelMsUser), nil
}

func (repo *CachedMsUserRepository) toView(u model.MsUser) MsUser {
	return MsUser{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Profile:  u.Profile,
		Username: u.Username,
		Birthday: time.UnixMilli(u.Birthday),
	}
}

func (repo *CachedMsUserRepository) toModel(u MsUser) model.MsUser {
	return model.MsUser{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.Birthday.UnixMilli(),
		Username: u.Username,
		Profile:  u.Profile,
	}
}

type MsUser struct {
	Id    int64
	Phone string
	Email string

	// encrypted password
	Password string

	Username string
	Profile  string

	Birthday time.Time
}
