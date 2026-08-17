package factory

import (
	"net/http/httptest"
	"testing"

	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

type accessTokenServiceStub struct {
	authentication.IAccessTokenService
}

func (*accessTokenServiceStub) Validate(token string) (string, bool, error) {
	return "user-1", token == "valid-token", nil
}

func TestProvideHttpVerifierResolvesTokenServiceFromScope(t *testing.T) {
	container := registry.NewContainer()
	registry.Export[authentication.IAccessTokenService](container, new(accessTokenServiceStub))
	scope := registry.NewScope()
	require.NoError(t, scope.Build(container))

	verifier := provideHttpVerifier(scope, TokenVerifier{HeaderKey: "X-Scene-Token"})
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Scene-Token", "valid-token")
	status, err := verifier.Verify(request)

	require.NoError(t, err)
	require.Equal(t, "user-1", status.UserID)
}
