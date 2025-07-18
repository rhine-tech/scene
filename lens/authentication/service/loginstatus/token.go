package loginstatus

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"net/http"
	"strings"
	"time"
)

// tokenAuth implements HTTPLoginStatusVerifier for token-based authentication.
type tokenAuth struct {
	srv       authentication.IAccessTokenService `aperture:""`
	headerKey string
	queryKey  string
}

// NewTokenAuth creates a new instance of a token-based verifier.
// if headerKey is not empty, it will also read token in header with that key
// if queryKey is not empty, it will also read token in query with specified key
func NewTokenAuth(srv authentication.IAccessTokenService,
	headerKey, queryKey string) authentication.HTTPLoginStatusVerifier {
	return &tokenAuth{srv: srv, headerKey: headerKey, queryKey: queryKey}
}

func (t *tokenAuth) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("HTTPLoginStatusVerifier", "access-token")
}

// Verify extracts a token from the "Authorization" header and validates it.
func (t *tokenAuth) Verify(request *http.Request) (authentication.LoginStatus, error) {
	var tokenValue string

	authHeader := request.Header.Get("Authorization")

	if authHeader != "" {
		// Expecting "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return authentication.LoginStatus{}, authentication.ErrNotLogin
		}
		tokenValue = parts[1]
	}

	if t.headerKey != "" && tokenValue == "" {
		tokenValue = request.Header.Get(t.headerKey)
	}

	if t.queryKey != "" && tokenValue == "" {
		tokenValue = request.URL.Query().Get(t.queryKey)
	}

	if tokenValue == "" {
		return authentication.LoginStatus{}, authentication.ErrNotLogin
	}

	// Validate the token using the service
	uid, valid, err := t.srv.Validate(tokenValue)
	if err != nil {
		return authentication.LoginStatus{}, authentication.ErrAuthenticationFailed
	}
	if !valid {
		return authentication.LoginStatus{}, authentication.ErrAuthenticationFailed
	}

	return authentication.LoginStatus{
		UserID:   uid,
		Verifier: t.SrvImplName().Implementation,
		Token:    tokenValue,
		ExpireAt: -1, // Expiry can be managed by the token service transparently
	}, nil
}

// Login creates a new token for the user. In a real-world scenario,
// the token would typically be returned in the response body.
func (t *tokenAuth) Login(userId string, resp http.ResponseWriter) (authentication.LoginStatus, error) {
	// Create a new token that expires in 24 hours
	expireAt := time.Now().Add(24 * time.Hour * 30).Unix()
	token, err := t.srv.Create(userId, "login-token", expireAt)
	if err != nil {
		return authentication.LoginStatus{}, err
	}

	// Note: The token should be sent to the client, usually in the response body.
	// Setting the header is an option but less common for login responses.
	// resp.Header().Set("Authorization", "Bearer "+token.Token)

	return authentication.LoginStatus{
		UserID:   userId,
		Verifier: t.SrvImplName().Implementation,
		Token:    token.Token,
		ExpireAt: -1,
	}, nil
}

// Logout invalidates the user's token.
// A proper implementation requires the token to be passed, e.g., via Verify() first.
// As the interface does not provide the request or token here, we assume
// the client will make a separate authenticated request to a 'delete token' endpoint.
func (t *tokenAuth) Logout(resp http.ResponseWriter) error {
	// To implement server-side logout, the token to be invalidated must be known.
	// This could be achieved by modifying the interface to accept *http.Request,
	// allowing token extraction from the header.
	// Example: `func (t *tokenAuth) Logout(request *http.Request, resp http.ResponseWriter) error`
	// Following the basicAuth example, this is a no-op.
	return nil
}
