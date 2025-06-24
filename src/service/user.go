package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/types/app"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Signup(ctx context.Context, u repository.User) error
	LoginWithEmailPwd(ctx context.Context, email string, password string) (repository.User, error)
	FindOrCreate(ctx context.Context, phone string) (repository.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo model.WechatInfo) (repository.User, error)
	EditProfile(ctx context.Context, u repository.User) (repository.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) Signup(ctx context.Context, u repository.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) LoginWithEmailPwd(ctx context.Context, email string, password string) (repository.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	switch err {
	case nil:
		// 检查密码对不对
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		if err != nil {
			return repository.User{}, app.ErrInvalidUserOrPassword
		}
		return u, nil
	case app.ErrRecordNotFound:
		return repository.User{}, app.ErrInvalidUserOrPassword
	default:
		return repository.User{}, err
	}
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (repository.User, error) {
	// 先找一下，我们认为，大部分用户是已经存在的用户
	u, err := svc.repo.FindByPhone(ctx, phone)
	switch err {
	case nil:
		return u, nil
	case app.ErrRecordNotFound:
		err = svc.repo.Create(ctx, repository.User{
			Phone: phone,
		})
		// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
		// 一种是 err != nil，系统错误
		if err != nil && !errors.Is(err, app.ErrDuplicateUser) {
			return repository.User{}, err
		}
		// 要么 err ==nil，要么ErrDuplicateUser，也代表用户存在
		// 主从延迟，理论上来讲，强制走主库
		return svc.repo.FindByPhone(ctx, phone)
	default:
		return repository.User{}, err
	}
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info model.WechatInfo) (repository.User, error) {
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if err != app.ErrRecordNotFound {
		return u, err
	}
	u = repository.User{
		WechaOpenId: info.OpenID,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && !errors.Is(err, app.ErrDuplicateUser) {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *userService) EditProfile(ctx context.Context, u repository.User) (repository.User, error) {
	_, err := svc.repo.FindById(ctx, u.Id)
	switch err {
	case nil:
		return svc.repo.EditProfile(ctx, u)

	case app.ErrRecordNotFound:
		return repository.User{}, app.ErrInvalidUserOrPassword

	default:
		return repository.User{}, err
	}
}
