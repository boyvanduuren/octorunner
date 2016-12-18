package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
	"golang.org/x/oauth2"
)

type Repository struct {
	Token, Secret string
}

type AuthMethod interface {
	RequestToken(repoFullName string) *oauth2.Token
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
