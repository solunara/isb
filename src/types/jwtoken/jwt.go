package jwtoken

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWT struct {
	SigningKey []byte
}

type CustomClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

func NewJWToken() *JWT {
	return &JWT{
		[]byte("XJvJWOadrbgUxUwqcXOKnnGpVwWPKgCA"),
	}
}

func NewClaims(username, encryptedPassword string) CustomClaims {
	return CustomClaims{
		username + "@" + encryptedPassword,
		jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000),
			ExpiresAt: int64(time.Now().Unix() + int64(72000)),
		},
	}
}

func (j *JWT) CreateJWToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParesJWToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		return nil, err
		// if ve, ok := err.(*jwt.ValidationError); ok {
		// 	if ve.Errors&jwt.ValidationErrorMalformed != 0 {
		// 		return nil, TokenMalformed
		// 	} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
		// 		Token is expired
		// 		return nil, TokenExpired
		// 	} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
		// 		return nil, TokenNotValidYet
		// 	} else {
		// 		return nil, TokenInvalid
		// 	}
		// }
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("couldn't handle this token")
}
