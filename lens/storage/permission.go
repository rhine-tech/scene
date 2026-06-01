package storage

import "github.com/rhine-tech/scene/lens/permission"

var (
	PermFileUpload   = permission.Create("storage:file:upload")
	PermFileDelete   = permission.Create("storage:file:delete")
	PermFileDownload = permission.Create("storage:file:download")
	PermFileList     = permission.Create("storage:file:list")
)
