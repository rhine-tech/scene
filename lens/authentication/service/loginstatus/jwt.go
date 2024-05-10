package loginstatus

import (
	"github.com/golang-jwt/jwt"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"net/http"
)

type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type JWT struct {
	secret  []byte
	keyName string
}

func NewJWT(secret []byte, keyName string) authentication.HTTPLoginStatusVerifier {
	return &JWT{secret: secret, keyName: keyName}
}

func (j *JWT) SrvImplName() scene.ImplName {
	return scene.NewSrvImplName("authentication", "HTTPLoginStatusVerifier", "jwt:cookie+header")
}

func (j *JWT) Verify(request *http.Request) (status authentication.LoginStatus, err error) {
	tokenStr := request.Header.Get(j.keyName)
	if tokenStr == "" {
		cookie, err := request.Cookie(j.keyName)
		if err != nil {
			return status, authentication.ErrNotLogin
		}
		tokenStr = cookie.Value
	}
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return authentication.LoginStatus{
			UserID:   claims.UserID,
			Token:    token.Raw,
			Verifier: j.SrvImplName().Version,
		}, nil
	} else {
		return status, err
	}
}

func (j *JWT) Login(userId string, resp http.ResponseWriter) (status authentication.LoginStatus, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		UserID: userId,
	})
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return authentication.LoginStatus{}, authentication.ErrNotLogin
	}
	http.SetCookie(resp, &http.Cookie{
		Name:  j.keyName,
		Value: tokenString,
		Path:  "/",
	})
	return authentication.LoginStatus{
		UserID:   userId,
		Token:    tokenString,
		Verifier: j.SrvImplName().Version,
		ExpireAt: -1,
	}, nil
}

func (j *JWT) Logout(resp http.ResponseWriter) (err error) {
	http.SetCookie(resp, &http.Cookie{
		Name:   j.keyName,
		Path:   "/",
		MaxAge: -1,
	})
	return nil
}
