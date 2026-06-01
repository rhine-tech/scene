package storage

import (
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// StorageKey is the unique identifier of a file in storage.
// it is composed with {Provider}://{ID}
// example: tos.buketName://objectName
// example: local.name://objectName
type StorageKey string

func NewStorageKey(provider string, path ...string) StorageKey {
	parts := make([]string, 0, len(path))
	for _, part := range path {
		if normalized := NormalizeIdentifier(part); normalized != "" {
			parts = append(parts, normalized)
		}
	}
	return StorageKey(provider + "://" + strings.Join(parts, "/"))
}

func NewStorageKeyWithUUID(provider string) StorageKey {
	return StorageKey(provider + "://" + strings.ReplaceAll(uuid.NewString(), "-", ""))
}

func ParseStorageKey(storageKey string) (StorageKey, bool) {
	key := StorageKey(storageKey)
	if ValidateStorageKey(key) != nil {
		return "", false
	}
	return key, true
}

func IsStorageKey(storageKey string) bool {
	return ValidateStorageKey(StorageKey(storageKey)) == nil
}

func (f StorageKey) Provider() string {
	return strings.Split(string(f), "://")[0]
}

func (f StorageKey) FileID() string {
	val := strings.Split(string(f), "://")
	if len(val) != 2 {
		return ""
	}
	return val[1]
}

func NormalizeIdentifier(identifier string) string {
	return strings.Trim(strings.TrimSpace(identifier), "/")
}

func ValidateStorageKey(storageKey StorageKey) error {
	parts := strings.Split(string(storageKey), "://")
	if len(parts) != 2 {
		return ErrInvalidStorageKey
	}
	if !validProviderName(parts[0]) {
		return ErrInvalidStorageKey
	}
	if !validIdentifier(parts[1]) {
		return ErrInvalidStorageKey
	}
	return nil
}

func validProviderName(provider string) bool {
	if provider == "" || strings.HasPrefix(provider, ".") || strings.HasSuffix(provider, ".") {
		return false
	}
	for _, part := range strings.Split(provider, ".") {
		if part == "" {
			return false
		}
		for _, r := range part {
			if !isProviderNameChar(r) {
				return false
			}
		}
	}
	return true
}

func isProviderNameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' ||
		r == '-'
}

func validIdentifier(identifier string) bool {
	if identifier == "" || identifier != NormalizeIdentifier(identifier) || strings.Contains(identifier, "\\") {
		return false
	}
	for _, r := range identifier {
		if unicode.IsControl(r) {
			return false
		}
	}
	for _, part := range strings.Split(identifier, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
