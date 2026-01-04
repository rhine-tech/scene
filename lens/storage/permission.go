package storage

import "github.com/rhine-tech/scene/lens/permission"

var (
	PermFileManage = permission.Create("storage:file:manage")
	PermFileUpload = permission.Create("storage:file:upload")
)
