package storage

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(5, "storage")

var (
	ErrStorageFailed             = _eg.CreateError(1, "storage failed")
	ErrFileNotFound              = _eg.CreateError(2, "file not found")
	ErrStorageNotFound           = _eg.CreateError(3, "storage not found")
	ErrFailToLoad                = _eg.CreateError(4, "fail to load")
	ErrFailToStore               = _eg.CreateError(5, "fail to store")
	ErrFailToDelete              = _eg.CreateError(6, "fail to delete")
	ErrLoadingMeta               = _eg.CreateError(7, "loading meta")
	ErrInvalidFileID             = _eg.CreateError(8, "invalid file id")
	ErrStorageError              = _eg.CreateError(9, "storage error")
	ErrInvalidOffset             = _eg.CreateError(10, "invalid offset")
	ErrInvalidLength             = _eg.CreateError(11, "invalid length")
	ErrUnknownError              = _eg.CreateError(12, "unknown error")
	ErrInitPartUploadFailed      = _eg.CreateError(13, "init part upload failed")
	ErrUploadSessionNotFound     = _eg.CreateError(14, "upload session not found")
	ErrStorePartFailed           = _eg.CreateError(15, "store part upload failed")
	ErrFailToAbortMultipartStore = _eg.CreateError(16, "fail to abort")
	ErrMetaNotFound              = _eg.CreateError(17, "meta not found")
	ErrFailToListMeta            = _eg.CreateError(18, "fail to list meta")
)
