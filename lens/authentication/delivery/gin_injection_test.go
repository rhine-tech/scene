package delivery

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/service/loginstatus"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

type authenticationServiceStub struct {
	authentication.IAuthenticationService
}

type storageServiceStub struct {
	storage.IStorageService
}

type loginTokenServiceStub struct {
	authentication.IAccessTokenService
}

func (*loginTokenServiceStub) Validate(token string) (string, bool, error) {
	return "user-1", token == "valid-token", nil
}

func TestAuthContextInjectsTokenVerifierService(t *testing.T) {
	tokenService := new(loginTokenServiceStub)
	verifier := loginstatus.NewTokenAuth(nil, "X-Scene-Token", "")
	context := &authContext{
		authSrv:  new(authenticationServiceStub),
		tokenSrv: tokenService,
		storage:  new(storageServiceStub),
		lgStVrf:  verifier,
	}
	container := registry.NewContainer()
	container.Load(context)
	container.Import(reflect.TypeFor[authentication.IAccessTokenService]().String(), tokenService)
	container.Inject()

	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Scene-Token", "valid-token")
	status, err := verifier.Verify(request)

	require.NoError(t, err)
	require.Equal(t, "user-1", status.UserID)
}
