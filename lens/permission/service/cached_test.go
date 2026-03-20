package service

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene"
	scache "github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/lens/permission"
)

type fakePermissionService struct {
	listCalls int
	addCalls  int
	delCalls  int
	data      map[string][]*permission.Permission
}

func newFakePermissionService() *fakePermissionService {
	return &fakePermissionService{
		data: map[string][]*permission.Permission{
			"u1": {permission.MustParsePermission("notify:send")},
		},
	}
}

func (f *fakePermissionService) SrvImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionService", "fake")
}

func (f *fakePermissionService) Setup() error { return nil }

func (f *fakePermissionService) HasPermission(owner string, perm *permission.Permission) bool {
	tree := permission.BuildTree(f.ListPermissions(owner)...)
	return tree.HasPermission(perm)
}

func (f *fakePermissionService) HasPermissionStr(owner string, perm string) bool {
	p, err := permission.ParsePermission(perm)
	if err != nil {
		return false
	}
	return f.HasPermission(owner, p)
}

func (f *fakePermissionService) ListPermissions(owner string) []*permission.Permission {
	f.listCalls++
	src := f.data[owner]
	out := make([]*permission.Permission, 0, len(src))
	for _, p := range src {
		out = append(out, p.Copy())
	}
	return out
}

func (f *fakePermissionService) AddPermission(owner string, perm string) error {
	f.addCalls++
	f.data[owner] = append(f.data[owner], permission.MustParsePermission(perm))
	return nil
}

func (f *fakePermissionService) RemovePermission(owner string, perm string) error {
	f.delCalls++
	current := f.data[owner]
	next := make([]*permission.Permission, 0, len(current))
	for _, p := range current {
		if p.String() != perm {
			next = append(next, p)
		}
	}
	f.data[owner] = next
	return nil
}

type fakeCache struct {
	items map[string][]byte
	tags  map[string]map[string]struct{}
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		items: make(map[string][]byte),
		tags:  make(map[string]map[string]struct{}),
	}
}

func (f *fakeCache) ImplName() scene.ImplName {
	return scache.Lens.ImplName("ICache", "fake")
}

func (f *fakeCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	v, ok := f.items[key]
	if !ok {
		return nil, false, nil
	}
	return v, true, nil
}

func (f *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration, tags ...string) error {
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

func (f *fakeCache) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		delete(f.items, key)
	}
	return nil
}

func (f *fakeCache) InvalidateTags(_ context.Context, tags ...string) error {
	for _, tag := range tags {
		for key := range f.tags[tag] {
			delete(f.items, key)
		}
		delete(f.tags, tag)
	}
	return nil
}

func TestCachedPermissionService_ListPermissionsCacheHit(t *testing.T) {
	base := newFakePermissionService()
	cached := NewCachedPermissionService(base)
	cached.CacheRepo = newFakeCache()
	if err := cached.Setup(); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	p1 := cached.ListPermissions("u1")
	p2 := cached.ListPermissions("u1")
	if base.listCalls != 1 {
		t.Fatalf("base list called %d times, want 1", base.listCalls)
	}
	if len(p1) != len(p2) || len(p1) != 1 || p1[0].String() != p2[0].String() {
		t.Fatal("cached permission list mismatch")
	}
}

func TestCachedPermissionService_InvalidateAfterAddRemove(t *testing.T) {
	base := newFakePermissionService()
	cached := NewCachedPermissionService(base)
	cached.CacheRepo = newFakeCache()
	if err := cached.Setup(); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_ = cached.ListPermissions("u1")
	if base.listCalls != 1 {
		t.Fatalf("base list called %d times, want 1", base.listCalls)
	}

	if err := cached.AddPermission("u1", "notify:receiver:list"); err != nil {
		t.Fatalf("add permission failed: %v", err)
	}
	_ = cached.ListPermissions("u1")
	if base.listCalls != 2 {
		t.Fatalf("after add, base list called %d times, want 2", base.listCalls)
	}

	if err := cached.RemovePermission("u1", "notify:receiver:list"); err != nil {
		t.Fatalf("remove permission failed: %v", err)
	}
	_ = cached.ListPermissions("u1")
	if base.listCalls != 3 {
		t.Fatalf("after remove, base list called %d times, want 3", base.listCalls)
	}
}

func TestCachedPermissionService_HasPermissionStrInvalid(t *testing.T) {
	base := newFakePermissionService()
	cached := NewCachedPermissionService(base)
	if cached.HasPermissionStr("u1", "notify::send") {
		t.Fatal("invalid permission string should return false")
	}
}
