package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
)

type Repository struct {
	Token, Secret string
}

type AuthMethod interface {
	RequestToken(repoFullName string) string
	RequestSecret(repoFullName string) []byte
}

func CalculateSignature(repoSecret []byte, payloadBody []byte) string {
	mac := hmac.New(sha1.New, repoSecret)
	mac.Write(payloadBody)
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func CompareSignatures(x []byte, y []byte) bool {
	if subtle.ConstantTimeCompare(x, y) == 1 {
		return true
	} else {
		return false
	}
}