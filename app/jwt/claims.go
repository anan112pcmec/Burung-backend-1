package jwt_function

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func ClaimsJWT(tokenString string) (id, email interface{}, err error) {

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing tidak sesuai: %v", t.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return "", "", fmt.Errorf("gagal parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["id"], claims["email"], nil
	}

	return "", "", fmt.Errorf("token tidak valid")
}
