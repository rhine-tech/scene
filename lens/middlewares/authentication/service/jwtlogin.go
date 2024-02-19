package service

import (
	"github.com/golang-jwt/jwt"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"net/http"
)

var jwtSecret = []byte("HFSAGF$GFASDGASDF$#@TGHS#@QRF")

type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type jwtAuthStatusService struct {
}

func NewJWTLoginStatusService() authentication.LoginStatusService {
	return &jwtAuthStatusService{}
}

func (j *jwtAuthStatusService) SrvImplName() scene.ImplName {
	return scene.NewSrvImplName("authentication", "LoginStatusService", "jwtAuthStatusService")
}

func (j *jwtAuthStatusService) Verify(request *http.Request) (status authentication.LoginStatus, err error) {
	tokenStr := request.Header.Get("jwt_token")
	if tokenStr == "" {
		cookie, err := request.Cookie("jwt_token")
		if err != nil {
			return status, authentication.ErrNotLogin
		}
		tokenStr = cookie.Value
	}
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return authentication.LoginStatus{UserID: claims.UserID, Token: token.Raw, Name: "jwt_token"}, nil
	} else {
		return status, err
	}
}

func (j *jwtAuthStatusService) Login(userId string, resp http.ResponseWriter) (status authentication.LoginStatus, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		UserID: userId,
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return authentication.LoginStatus{}, authentication.ErrNotLogin
	}
	http.SetCookie(resp, &http.Cookie{
		Name:  "jwt_token",
		Value: tokenString,
		Path:  "/",
	})
	return authentication.LoginStatus{
		UserID:   userId,
		Token:    tokenString,
		Name:     "jwt_token",
		ExpireAt: -1,
	}, nil
}

func (j *jwtAuthStatusService) Logout(resp http.ResponseWriter) (err error) {
	http.SetCookie(resp, &http.Cookie{
		Name:   "jwt_token",
		Path:   "/",
		MaxAge: -1,
	})
	return nil
}
