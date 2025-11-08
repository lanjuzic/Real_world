package jwt

import (
	"kratos-realworld/internal/conf"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
)

type CustomClaims struct { //自定义断言
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
}

func NewJWTService(c *conf.Auth) *JWTService {
	return &JWTService{
		secret: []byte(c.JwtSecret),
	}
}

func NewConfAuth() *conf.Auth {
	return &conf.Auth{}
}

//var secretKey = []byte("your-secret-key")

func (j *JWTService) GenerateToken(userID int64, email string) (string, error) { //签发token
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "kratos-realworld",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secret)
}

func (j *JWTService) ParseToken(tokenStr string) (*CustomClaims, error) { //解析token展示用不上
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

var ProviderSet = wire.NewSet(NewJWTService)
