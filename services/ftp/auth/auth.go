package auth

import "SimPro/pkg/goftp.io/server/v2"

// Authentication scheme that accepts each user/password combination
type ZeroAuth struct{}

func (i *ZeroAuth) CheckPasswd(ctx *server.Context, u string, p string) (bool, error) {
	return true, nil
}

// Authentication scheme
type UserPass struct {
	User  string
	Pass  string
	IsSet bool
}

func (u *UserPass) GetIsSet() bool      { return u.IsSet }
func (u *UserPass) GetUser() string     { return u.User }
func (u *UserPass) GetPassword() string { return u.Pass }
