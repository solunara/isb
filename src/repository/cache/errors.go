package cache

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSystemError = errors.New("system error")

	ErrWrongCode = errors.New("wrong captche")

	ErrCodeVerifyTooManyTimes = errors.New("code verified too many times")

	ErrSendTooFrequently = errors.New("send too frequently")

	ErrUnknownCode = errors.New("unknown for code")

	ErrKeyNotExist = redis.Nil
)
