package service

import (
	"context"

	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/types/app"
	"golang.org/x/crypto/bcrypt"
)

type MsUserService interface {
	Signup(ctx context.Context, u repository.MsUser) error
	LoginWithPwd(ctx context.Context, username string, password string) (repository.MsUser, error)
}

type msUserService struct {
	repo repository.MsUserRepository
}

func NewMsUserService(repo repository.MsUserRepository) MsUserService {
	return &msUserService{
		repo: repo,
	}
}

func (svc *msUserService) Signup(ctx context.Context, u repository.MsUser) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *msUserService) LoginWithPwd(ctx context.Context, username string, password string) (repository.MsUser, error) {
	u, err := svc.repo.FindByUsername(ctx, username)
	switch err {
	case nil:
		// 检查密码对不对
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		if err != nil {
			return repository.MsUser{}, app.ErrInvalidUserOrPassword
		}
		return u, nil
	case app.ErrRecordNotFound:
		return repository.MsUser{}, app.ErrInvalidUserOrPassword
	default:
		return repository.MsUser{}, err
	}
}

// func (svc *msUserService) FindOrCreate(ctx context.Context, phone string) (repository.User, error) {
// 	// 先找一下，我们认为，大部分用户是已经存在的用户
// 	u, err := svc.repo.FindByPhone(ctx, phone)
// 	switch err {
// 	case nil:
// 		return u, nil
// 	case app.ErrRecordNotFound:
// 		err = svc.repo.Create(ctx, repository.User{
// 			Phone: phone,
// 		})
// 		// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
// 		// 一种是 err != nil，系统错误
// 		if err != nil && !errors.Is(err, app.ErrDuplicateUser) {
// 			return repository.User{}, err
// 		}
// 		// 要么 err ==nil，要么ErrDuplicateUser，也代表用户存在
// 		// 主从延迟，理论上来讲，强制走主库
// 		return svc.repo.FindByPhone(ctx, phone)
// 	default:
// 		return repository.User{}, err
// 	}
// }
