package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene"
	scache "github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
)

type stubLogger struct{}

func (stubLogger) Debug(args ...interface{})                           {}
func (stubLogger) Debugf(format string, args ...interface{})           {}
func (stubLogger) DebugW(message string, keysAndValues ...interface{}) {}
func (stubLogger) DebugS(message string, fields logger.LogField)       {}
func (stubLogger) Info(args ...interface{})                            {}
func (stubLogger) Infof(format string, args ...interface{})            {}
func (stubLogger) InfoW(message string, keysAndValues ...interface{})  {}
func (stubLogger) InfoS(message string, fields logger.LogField)        {}
func (stubLogger) Warn(args ...interface{})                            {}
func (stubLogger) Warnf(format string, args ...interface{})            {}
func (stubLogger) WarnW(message string, keysAndValues ...interface{})  {}
func (stubLogger) WarnS(message string, fields logger.LogField)        {}
func (stubLogger) Error(args ...interface{})                           {}
func (stubLogger) Errorf(format string, args ...interface{})           {}
func (stubLogger) ErrorW(message string, keysAndValues ...interface{}) {}
func (stubLogger) ErrorS(message string, fields logger.LogField)       {}
func (stubLogger) WithPrefix(prefix string) logger.ILogger             { return stubLogger{} }
func (stubLogger) SetLogLevel(level logger.LogLevel)                   {}
func (stubLogger) WithOptions(opts ...logger.Option) logger.ILogger    { return stubLogger{} }

type fakeAuthenticationService struct {
	userByIDCalls   int
	userByNameCalls int
	listCalls       int
	usersByID       map[string]authentication.User
	usersByName     map[string]authentication.User
}

func newFakeAuthenticationService() *fakeAuthenticationService {
	user := authentication.User{
		UserID:   "u1",
		Username: "alice",
		Email:    "alice@example.com",
	}
	return &fakeAuthenticationService{
		usersByID: map[string]authentication.User{
			user.UserID: user,
		},
		usersByName: map[string]authentication.User{
			user.Username: user,
		},
	}
}

func (f *fakeAuthenticationService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationService", "fake")
}

func (f *fakeAuthenticationService) Setup() error { return nil }
func (f *fakeAuthenticationService) AddUser(username, password string) (authentication.User, error) {
	panic("not used")
}
func (f *fakeAuthenticationService) DeleteUser(userID string) error { panic("not used") }
func (f *fakeAuthenticationService) Authenticate(username string, password string) (string, error) {
	panic("not used")
}
func (f *fakeAuthenticationService) AuthenticateByToken(token string) (string, error) {
	panic("not used")
}
func (f *fakeAuthenticationService) HasUser(userID string) (bool, error) { panic("not used") }

func (f *fakeAuthenticationService) UpdateUser(user authentication.User) error {
	prev, ok := f.usersByID[user.UserID]
	if ok {
		delete(f.usersByName, prev.Username)
	}
	f.usersByID[user.UserID] = user
	f.usersByName[user.Username] = user
	return nil
}

func (f *fakeAuthenticationService) UserById(userID string) (authentication.User, error) {
	f.userByIDCalls++
	user, ok := f.usersByID[userID]
	if !ok {
		return authentication.User{}, authentication.ErrUserNotFound
	}
	return user, nil
}

func (f *fakeAuthenticationService) UserByName(username string) (authentication.User, error) {
	f.userByNameCalls++
	user, ok := f.usersByName[username]
	if !ok {
		return authentication.User{}, authentication.ErrUserNotFound
	}
	return user, nil
}

func (f *fakeAuthenticationService) UserByEmail(email string) (authentication.User, error) {
	panic("not used")
}

func (f *fakeAuthenticationService) ListUsers(offset, limit int64) (model.PaginationResult[authentication.User], error) {
	f.listCalls++
	results := make([]authentication.User, 0, len(f.usersByID))
	for _, user := range f.usersByID {
		results = append(results, user)
	}
	return model.PaginationResult[authentication.User]{
		Total:   int64(len(results)),
		Offset:  offset,
		Count:   int64(len(results)),
		Results: results,
	}, nil
}

type fakeAuthCache struct {
	items map[string][]byte
	tags  map[string]map[string]struct{}
}

func newFakeAuthCache() *fakeAuthCache {
	return &fakeAuthCache{
		items: make(map[string][]byte),
		tags:  make(map[string]map[string]struct{}),
	}
}

func (f *fakeAuthCache) ImplName() scene.ImplName {
	return scache.Lens.ImplName("ICache", "fake")
}

func (f *fakeAuthCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	v, ok := f.items[key]
	if !ok {
		return nil, false, nil
	}
	return v, true, nil
}

func (f *fakeAuthCache) Set(_ context.Context, key string, value []byte, _ time.Duration, tags ...string) error {
	f.items[key] = append([]byte(nil), value...)
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if f.tags[tag] == nil {
			f.tags[tag] = make(map[string]struct{})
		}
		f.tags[tag][key] = struct{}{}
	}
	return nil
}

func (f *fakeAuthCache) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		delete(f.items, key)
	}
	return nil
}

func (f *fakeAuthCache) InvalidateTags(_ context.Context, tags ...string) error {
	for _, tag := range tags {
		for key := range f.tags[tag] {
			delete(f.items, key)
		}
		delete(f.tags, tag)
	}
	return nil
}

func newCachedAuthenticationServiceForTest(base authentication.IAuthenticationService) *CachedAuthenticationService {
	cached := NewCachedAuthenticationService(base).(*CachedAuthenticationService)
	cached.cache = newFakeAuthCache()
	cached.log = stubLogger{}
	if err := cached.Setup(); err != nil {
		panic(err)
	}
	return cached
}

func TestCachedAuthenticationService_UserByIDCacheHit(t *testing.T) {
	base := newFakeAuthenticationService()
	cached := newCachedAuthenticationServiceForTest(base)

	user1, err := cached.UserById("u1")
	if err != nil {
		t.Fatalf("first lookup failed: %v", err)
	}
	user2, err := cached.UserById("u1")
	if err != nil {
		t.Fatalf("second lookup failed: %v", err)
	}
	if base.userByIDCalls != 1 {
		t.Fatalf("user by id called %d times, want 1", base.userByIDCalls)
	}
	if user1.UserID != user2.UserID || user2.Username != "alice" {
		t.Fatalf("unexpected cached user payload: %+v %+v", user1, user2)
	}
}

func TestCachedAuthenticationService_InvalidateUserCacheAfterUpdate(t *testing.T) {
	base := newFakeAuthenticationService()
	cached := newCachedAuthenticationServiceForTest(base)

	if _, err := cached.UserByName("alice"); err != nil {
		t.Fatalf("initial lookup failed: %v", err)
	}
	if base.userByNameCalls != 1 {
		t.Fatalf("user by name called %d times, want 1", base.userByNameCalls)
	}

	err := cached.UpdateUser(authentication.User{
		UserID:   "u1",
		Username: "alice-renamed",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if _, err = cached.UserByName("alice-renamed"); err != nil {
		t.Fatalf("lookup after update failed: %v", err)
	}
	if base.userByNameCalls != 2 {
		t.Fatalf("user by name called %d times after invalidation, want 2", base.userByNameCalls)
	}
}

func TestCachedAuthenticationService_InvalidateListCacheAfterUpdate(t *testing.T) {
	base := newFakeAuthenticationService()
	cached := newCachedAuthenticationServiceForTest(base)

	if _, err := cached.ListUsers(0, 20); err != nil {
		t.Fatalf("initial list failed: %v", err)
	}
	if _, err := cached.ListUsers(0, 20); err != nil {
		t.Fatalf("cached list failed: %v", err)
	}
	if base.listCalls != 1 {
		t.Fatalf("list called %d times, want 1", base.listCalls)
	}

	err := cached.UpdateUser(authentication.User{
		UserID:   "u1",
		Username: "alice-v2",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	result, err := cached.ListUsers(0, 20)
	if err != nil {
		t.Fatalf("list after update failed: %v", err)
	}
	if base.listCalls != 2 {
		t.Fatalf("list called %d times after invalidation, want 2", base.listCalls)
	}
	if len(result.Results) != 1 || result.Results[0].Username != "alice-v2" {
		t.Fatalf("unexpected list result after update: %+v", result.Results)
	}
}
