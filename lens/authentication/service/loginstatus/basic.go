package loginstatus

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"net/http"
)

type basicAuth struct {
	srv authentication.AuthenticationService `aperture:""`
}

func NewBasicAuth(srv authentication.AuthenticationService) authentication.HTTPLoginStatusVerifier {
	return &basicAuth{srv: srv}
}

func (b *basicAuth) SrvImplName() scene.ImplName {
	return scene.NewSrvImplName("authentication", "HTTPLoginStatusVerifier", "basic")
}

func (b *basicAuth) Verify(request *http.Request) (status authentication.LoginStatus, err error) {
	user, password, ok := request.BasicAuth()
	if !ok {
		return status, authentication.ErrNotLogin
	}
	uid, err := b.srv.Authenticate(user, password)
	if err != nil {
		return status, err
	}
	return authentication.LoginStatus{
		UserID:   uid,
		Verifier: b.SrvImplName().Implementation,
		Token:    "",
		ExpireAt: -1,
	}, nil
}

func (b *basicAuth) Login(userId string, resp http.ResponseWriter) (status authentication.LoginStatus, err error) {
	return authentication.LoginStatus{
		UserID:   userId,
		Verifier: b.SrvImplName().Implementation,
		Token:    "",
		ExpireAt: -1,
	}, nil
}

func (b *basicAuth) Logout(resp http.ResponseWriter) (err error) {
	return nil
}
