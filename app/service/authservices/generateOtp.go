package authservices

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP() string {
	otp := ""
	for i := 0; i < 8; i++ {
		// ambil angka random 0–9
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		otp += fmt.Sprintf("%d", n.Int64())
	}
	return otp
}
