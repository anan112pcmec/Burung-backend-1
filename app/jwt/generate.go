package jwt_function

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("Rahasia")

func GenerateJWT(id int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"id":    id,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(), // expire 1 hari
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		fmt.Println("Gagal Membuat Jwt")
		return "", fmt.Errorf("Gagal Membuat Jwt")
	}

	return signedToken, nil
}
