package auth

import (
	"golang.org/x/oauth2"
)

type SimpleAuth struct {
	Store map[string]Repository
}

func (auth SimpleAuth) RequestToken(repoFullName string) *oauth2.Token {
	var token *oauth2.Token
	if val, exists := auth.Store[repoFullName]; exists {
		token = &oauth2.Token{AccessToken: val.Token}
	}
	return token
}

func (auth SimpleAuth) RequestSecret(repoFullName string) []byte {
	var secret []byte
	if val, exists := auth.Store[repoFullName]; exists {
		secret = []byte(val.Secret)
	}
	return secret
}
