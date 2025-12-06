package test

import (
	"time"

	"github.com/pquerna/otp/totp"
)

func GenerateTOTP(secret string) string {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		panic(err)
	}
	return code
}