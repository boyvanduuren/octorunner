package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
	"golang.org/x/oauth2"
)

/*
Repository is used to store tokens and secrets per repository.
Tokens are used for downloading private repositories, setting statuses, etc. Secret
are used to verify clients.
*/
type Repository struct {
	Token, Secret string
}

/*
Method is a simple interface every authentication method has to implement. This might change in the future.
*/
type Method interface {
	RequestToken(repoFullName string) *oauth2.Token
	RequestSecret(repoFullName string) []byte
}

/*
CalculateSignature calculates a signature based on a (repository's) secret.
*/
func CalculateSignature(repoSecret []byte, payloadBody []byte) string {
	mac := hmac.New(sha1.New, repoSecret)
	mac.Write(payloadBody)
	return fmt.Sprintf("%x", mac.Sum(nil))
}

/*
CompareSignatures does a constant time comparison of two signatures.
*/
func CompareSignatures(x []byte, y []byte) bool {
	if subtle.ConstantTimeCompare(x, y) == 1 {
		return true
	}
	return false
}
