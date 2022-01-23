package security

import (
	"errors"
	"fmt"
	"site/database/models"

	"github.com/golang-jwt/jwt"
)

var hmacSampleSecret []byte

type TokenSecurity interface {
	CreateToken(u *models.User) (string, error)
	IsValid(token string) bool
	GetIdentifier(token string) (uint, error)
}

func NewTokenService() *TokenService {
	return &TokenService{}
}

type TokenService struct {
}

func (t TokenService) CreateToken(u *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": u.ID,
	})

	return token.SignedString(hmacSampleSecret)
}

func (t TokenService) IsValid(token string) bool {
	jwtT, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	if err != nil {
		return false
	}

	// ok && jwtT.Valid
	_, ok := jwtT.Claims.(jwt.MapClaims)
	if !ok || !jwtT.Valid {
		return false
	}

	return true

}

func (t TokenService) GetIdentifier(token string) (uint, error) {
	jwtT, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	if err != nil {
		return 0, errors.New("invalid Token")
	}

	// ok && jwtT.Valid
	claims, ok := jwtT.Claims.(jwt.MapClaims)
	if !ok || !jwtT.Valid {
		return 0, errors.New("invalid Token")
	}

	id := uint(claims["id"].(float64))

	return id, nil
}
