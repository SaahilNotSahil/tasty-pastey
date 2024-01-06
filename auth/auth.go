package auth

import (
	"errors"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Jwt struct {
	SecretKey         string
	Issuer            string
	ExpirationMinutes int64
	ExpirationHours   int64
}

type JwtClaim struct {
	Email string
	jwt.StandardClaims
}

func (j *Jwt) GenerateToken(email string) (signedToken string, err error) {
	claims := &JwtClaim{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(
				time.Minute * time.Duration(j.ExpirationMinutes),
			).Unix(),
			Issuer: j.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err = token.SignedString([]byte(j.SecretKey))
	if err != nil {
		log.Println(err)

		return "", err
	}

	return signedToken, nil
}

func (j *Jwt) RefreshToken(email string) (signedtoken string, err error) {
	claims := &JwtClaim{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(
				time.Hour * time.Duration(j.ExpirationHours),
			).Unix(),
			Issuer: j.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedtoken, err = token.SignedString([]byte(j.SecretKey))
	if err != nil {
		log.Println(err)

		return "", err
	}

	return signedtoken, nil
}

func (j *Jwt) ValidateToken(signedToken string) (claims *JwtClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.SecretKey), nil
		},
	)
	if err != nil {
		log.Println(err)

		return
	}

	claims, ok := token.Claims.(*JwtClaim)
	if !ok {
		err = errors.New("couldn't parse claims")

		log.Println(err)

		return nil, err
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("JWT is expired")

		log.Println(err)

		return nil, err
	}

	return claims, nil
}
