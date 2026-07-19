package base

import (
	"context"
	"errors"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
)

type stubLogger struct{}

func (stubLogger) Debug(args ...interface{})                           {}
func (stubLogger) Debugf(format string, args ...interface{})           {}
func (stubLogger) DebugW(message string, keysAndValues ...interface{}) {}
func (stubLogger) Info(args ...interface{})                            {}
func (stubLogger) Infof(format string, args ...interface{})            {}
func (stubLogger) InfoW(message string, keysAndValues ...interface{})  {}
func (stubLogger) Warn(args ...interface{})                            {}
func (stubLogger) Warnf(format string, args ...interface{})            {}
func (stubLogger) WarnW(message string, keysAndValues ...interface{})  {}
func (stubLogger) Error(args ...interface{})                           {}
func (stubLogger) Errorf(format string, args ...interface{})           {}
func (stubLogger) ErrorW(message string, keysAndValues ...interface{}) {}
func (stubLogger) WithPrefix(prefix string) logger.ILogger             { return stubLogger{} }
func (stubLogger) SetLogLevel(level logger.LogLevel)                   {}
func (stubLogger) WithOptions(opts ...logger.Option) logger.ILogger    { return stubLogger{} }

type fakeAuthRepo struct {
	userByNameFn func(username string) (authentication.User, error)
	addUserFn    func(user authentication.User) (authentication.User, error)
}

func (f *fakeAuthRepo) ImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationRepository", "fake")
}

func (f *fakeAuthRepo) Authenticate(context.Context, string, string) (string, error) {
	panic("not used")
}

func (f *fakeAuthRepo) UserById(context.Context, string) (authentication.User, error) {
	panic("not used")
}

func (f *fakeAuthRepo) UserByName(_ context.Context, username string) (authentication.User, error) {
	return f.userByNameFn(username)
}

func (f *fakeAuthRepo) UserByEmail(context.Context, string) (authentication.User, error) {
	panic("not used")
}

func (f *fakeAuthRepo) AddUser(_ context.Context, user authentication.User) (authentication.User, error) {
	return f.addUserFn(user)
}

func (f *fakeAuthRepo) DeleteUser(context.Context, string) error {
	panic("not used")
}

func (f *fakeAuthRepo) UpdateUser(context.Context, authentication.User) error {
	panic("not used")
}

func (f *fakeAuthRepo) ListUsers(
	context.Context,
	int64,
	int64,
) (model.PaginationResult[authentication.User], error) {
	panic("not used")
}

func newTestService(repo authentication.IAuthenticationRepository) *authenticationService {
	return &authenticationService{
		logger:   stubLogger{},
		userRepo: repo,
	}
}

func TestAddUser_NewUser_Success(t *testing.T) {
	repo := &fakeAuthRepo{
		userByNameFn: func(username string) (authentication.User, error) {
			return authentication.User{}, authentication.ErrUserNotFound
		},
		addUserFn: func(user authentication.User) (authentication.User, error) {
			if user.UserID == "" {
				t.Fatalf("expected generated user id")
			}
			if user.Username != "alice" || user.Password != "secret" {
				t.Fatalf("unexpected user payload: %+v", user)
			}
			return user, nil
		},
	}

	got, err := newTestService(repo).AddUser("alice", "secret")
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("unexpected username: %s", got.Username)
	}
	if got.UserID == "" {
		t.Fatalf("expected non-empty user id")
	}
}

func TestAddUser_DuplicateUser_ReturnsAlreadyExists(t *testing.T) {
	repo := &fakeAuthRepo{
		userByNameFn: func(username string) (authentication.User, error) {
			return authentication.User{UserID: "u1", Username: username}, nil
		},
		addUserFn: func(user authentication.User) (authentication.User, error) {
			t.Fatalf("AddUser should not be called for duplicate user")
			return authentication.User{}, nil
		},
	}

	_, err := newTestService(repo).AddUser("alice", "secret")
	if !errors.Is(err, authentication.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestAddUser_UserLookupRepositoryError_ReturnsFailToAddUser(t *testing.T) {
	repoErr := errors.New("db unavailable")
	repo := &fakeAuthRepo{
		userByNameFn: func(username string) (authentication.User, error) {
			return authentication.User{}, repoErr
		},
		addUserFn: func(user authentication.User) (authentication.User, error) {
			t.Fatalf("AddUser should not be called when lookup fails")
			return authentication.User{}, nil
		},
	}

	_, err := newTestService(repo).AddUser("alice", "secret")
	if !errors.Is(err, authentication.ErrFailToAddUser) {
		t.Fatalf("expected ErrFailToAddUser, got: %v", err)
	}
}
