package repository

import (
	"context"
	"database/sql"
	"github.com/solunara/isb/src/types/app"
	"log"
	"time"

	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/repository/cache"
	"github.com/solunara/isb/src/repository/dao"
)

type UserRepository interface {
	Create(ctx context.Context, u User) error
	FindById(ctx context.Context, uid int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	EditProfile(ctx context.Context, u User) (User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, u User) error {
	return repo.dao.Insert(ctx, repo.toModel(u))
}

func (repo *CachedUserRepository) FindById(ctx context.Context, uid int64) (User, error) {
	modelUser, err := repo.cache.Get(ctx, uid)
	switch err {
	case nil:
		// 只要 err 为 nil，就返回
		return repo.toView(modelUser), nil
	case app.ErrKeyNotExist:
		modelUser, err = repo.dao.FindById(ctx, uid)
		if err != nil {
			return User{}, err
		}
		//modelUser = repo.toModel(u)
		//go func() {
		//	err = repo.cache.Set(ctx, du)
		//	if err != nil {
		//		log.Println(err)
		//	}
		//}()

		err = repo.cache.Set(ctx, modelUser, time.Hour*240)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
			log.Println(err)
		}
		return repo.toView(modelUser), nil
	default:
		// 接近降级的写法
		return User{}, err
	}
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	modelUser, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	return repo.toView(modelUser), nil
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (User, error) {
	modelUser, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return User{}, err
	}
	return repo.toView(modelUser), nil
}

func (repo *CachedUserRepository) EditProfile(ctx context.Context, u User) (User, error) {
	modelUser, err := repo.dao.UpdateUser(ctx, repo.toModel(u))
	if err != nil {
		return User{}, err
	}
	return repo.toView(modelUser), nil
}

func (repo *CachedUserRepository) toView(u model.User) User {
	return User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Profile:  u.Profile,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
	}
}

func (repo *CachedUserRepository) toModel(u User) model.User {
	return model.User{
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
		Nickname: u.Nickname,
		Profile:  u.Profile,
	}
}

type User struct {
	Id    int64
	Phone string
	Email string

	// encrypted password
	Password string

	Nickname string
	Profile  string

	Birthday time.Time
}
