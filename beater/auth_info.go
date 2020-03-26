package beater

import (
	"fmt"
	"time"
)

type authInfo struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	setTime      time.Time
}

func (a *authInfo) header() string {
	return fmt.Sprintf("%s %s", a.TokenType, a.AccessToken)
}

func (a *authInfo) expired() bool {
	const expirationBuffer = 60 // extra seconds unexpired token is considered expired
	return time.Now().Unix()-a.setTime.Unix() > int64(a.ExpiresIn-expirationBuffer)
}
