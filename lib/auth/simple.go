package auth

type SimpleAuth struct {
	Store map[string]Repository
}

func (auth SimpleAuth) RequestToken(repoFullName string) string {
	var token string
	if val, exists := auth.Store[repoFullName]; exists {
		token = val.Token
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
