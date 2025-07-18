package authentication

import "github.com/rhine-tech/scene/lens/permission"

var (
	PermAdmin       = permission.MustParsePermission("authentication:admin")
	PermTokenCreate = permission.MustParsePermission("authentication:token:create")
	PermTokenList   = permission.MustParsePermission("authentication:token:list")
	PermTokenDelete = permission.MustParsePermission("authentication:token:delete")
)
