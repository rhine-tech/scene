package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateStorageKey(t *testing.T) {
	tests := []struct {
		name string
		key  StorageKey
		ok   bool
	}{
		{name: "valid nested identifier", key: "local.default://covers/work-1/file.jpg", ok: true},
		{name: "valid generated key", key: NewStorageKeyWithUUID("s3.default"), ok: true},
		{name: "missing separator", key: "local.default/file.jpg", ok: false},
		{name: "empty provider", key: "://file.jpg", ok: false},
		{name: "empty identifier", key: "local.default://", ok: false},
		{name: "invalid provider character", key: "local default://file.jpg", ok: false},
		{name: "empty identifier segment", key: "local.default://a//b", ok: false},
		{name: "dot segment", key: "local.default://a/./b", ok: false},
		{name: "parent segment", key: "local.default://a/../b", ok: false},
		{name: "leading slash", key: "local.default:///a/b", ok: false},
		{name: "backslash", key: "local.default://a\\b", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStorageKey(tt.key)
			if tt.ok {
				require.NoError(t, err)
				require.True(t, IsStorageKey(string(tt.key)))
				parsed, ok := ParseStorageKey(string(tt.key))
				require.True(t, ok)
				require.Equal(t, tt.key, parsed)
				return
			}
			require.ErrorIs(t, err, ErrInvalidStorageKey)
			require.False(t, IsStorageKey(string(tt.key)))
			_, ok := ParseStorageKey(string(tt.key))
			require.False(t, ok)
		})
	}
}

func TestNewStorageKeyNormalizesIdentifier(t *testing.T) {
	key := NewStorageKey("local.default", "/covers/", "work-1.jpg/")
	require.Equal(t, StorageKey("local.default://covers/work-1.jpg"), key)
	require.NoError(t, ValidateStorageKey(key))
}

func TestValidateStorageKeyReturnsModuleError(t *testing.T) {
	err := ValidateStorageKey("bad")
	require.True(t, errors.Is(err, ErrInvalidStorageKey))
}
