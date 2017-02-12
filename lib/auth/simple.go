package auth

import (
	"golang.org/x/oauth2"
)

/*
SimpleAuth is a map from strings representing repository names to its token and/or secret.
*/
type SimpleAuth struct {
	Store map[string]Repository
}

/*
RequestToken the token that belongs to a specific repository.
*/
func (auth SimpleAuth) RequestToken(repoFullName string) *oauth2.Token {
	var token *oauth2.Token
	if val, exists := auth.Store[repoFullName]; exists {
		token = &oauth2.Token{AccessToken: val.Token}
	}
	return token
}

/*
RequestSecret returns the secret that belongs to a specific repository.
*/
func (auth SimpleAuth) RequestSecret(repoFullName string) []byte {
	var secret []byte
	if val, exists := auth.Store[repoFullName]; exists {
		secret = []byte(val.Secret)
	}
	return secret
}
