package app

import (
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// cache err
const (
	ErrKeyNotExist = redis.Nil
)

// dao err
var (
	ErrDuplicateEmail = errors.New("email address conflict")
	ErrRecordNotFound = gorm.ErrRecordNotFound
	ErrDuplicateUser  = gorm.ErrDuplicatedKey
)

// service err
var (
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

// 预定义错误码
const (
	ErrCodeInternalServer = 500
	ErrCodeBadRequest     = 400
	ErrCodeUnauthorized   = 401
	ErrCodeForbidden      = 403
	ErrCodeNotFound       = 404
	ErrCodeConflict       = 409
)

// web 预定义错误, 可直接返回给客户端
var (
	ErrInternalServer = &ResponseType{
		Code: ErrCodeInternalServer,
		Msg:  "服务器内部错误",
	}

	ErrBadRequest = &ResponseType{
		Code: ErrCodeBadRequest,
		Msg:  "请求参数格式错误",
	}

	ErrBadRequestInvalidEmail = &ResponseType{
		Code: ErrCodeBadRequest,
		Msg:  "邮箱格式错误",
	}

	ErrBadRequestInvalidPassword = &ResponseType{
		Code: ErrCodeBadRequest,
		Msg:  "the password must contain letters, numbers, special characters and not less than 8 digits",
	}

	ErrBadRequestWrongPassword = &ResponseType{
		Code: ErrCodeBadRequest,
		Msg:  "密码错误",
	}

	ErrBadRequestWrongBirthday = &ResponseType{
		Code: ErrCodeBadRequest,
		Msg:  "生日格式错误",
	}

	ErrBadRequestErrInvalidUserOrPassword = &ResponseType{
		Code: ErrCodeBadRequest,
		Msg:  ErrInvalidUserOrPassword.Error(),
	}

	ErrUnauthorized = &ResponseType{
		Code: ErrCodeUnauthorized,
		Msg:  "未授权访问",
	}

	ErrForbidden = &ResponseType{
		Code: ErrCodeForbidden,
		Msg:  "禁止访问",
	}

	ErrNotFound = &ResponseType{
		Code: ErrCodeNotFound,
		Msg:  "资源不存在",
	}

	ErrConflict = &ResponseType{
		Code: ErrCodeConflict,
		Msg:  "资源冲突",
	}
)
