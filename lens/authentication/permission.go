package authentication

import "github.com/rhine-tech/scene/lens/permission"

var (
	PermAdmin       = permission.Create("authentication:admin")
	PermTokenCreate = permission.Create("authentication:token:create")
	PermTokenList   = permission.Create("authentication:token:list")
	PermTokenDelete = permission.Create("authentication:token:delete")
)
