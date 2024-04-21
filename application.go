package scene

import "fmt"

type AppStatus int

const (
	AppStatusStopped AppStatus = iota
	AppStatusRunning
	AppStatusError
)

type Application interface {
	Name() ImplName // return scene
	//Status() AppStatus
	//Error() error
}

type Repository interface {
	RepoImplName() ImplName
}

type Service interface {
	SrvImplName() ImplName
}

/*
ImplName is the name of an implementation.
It is used to identify an implementation of a repository, service or application.
*/

type ImplType string

const (
	ImplTypeCore   = ImplType("core")   // core type
	ImplTypeScene  = ImplType("scene")  // scenario type
	ImplTypeInfra  = ImplType("infra")  // infrastructure type
	ImplTypeModule = ImplType("module") // module type
	ImplTypeRepo   = ImplType("repo")   // module repository type
	ImplTypeSrv    = ImplType("srv")    // module service type
	ImplTypeApp    = ImplType("app")    // module application type
)

type NamableImplementation interface {
	ImplName() ImplName
}

type ImplName struct {
	ImplType       ImplType
	Module         string
	Implementation string
	Version        string
}

// String returns a string representation of the implementation name.
// String is a pretty representation of the implementation name.
// If you want to use the implementation name as an identifier, use Identifier().
func (i ImplName) String() string {
	if i.Version == "" {
		return fmt.Sprintf("(%s)%s:%s", string(i.ImplType), i.Module, i.Implementation)
	}
	return fmt.Sprintf("(%s)%s:%s:%s", string(i.ImplType), i.Module, i.Implementation, i.Version)
}

func (i ImplName) EndpointName() string {
	return i.Module + "/" + i.Implementation
}

// Identifier returns a string identifier of the implementation name.
func (i ImplName) Identifier() string {
	return fmt.Sprintf("(%s)%s:%s:%s", string(i.ImplType), i.Module, i.Implementation, i.Version)
}

func NewImplName(implType ImplType, module, implementation, version string) ImplName {
	return ImplName{
		ImplType:       implType,
		Module:         module,
		Implementation: implementation,
		Version:        version,
	}
}

func NewImplNameNoVer(implType ImplType, module, implementation string) ImplName {
	return ImplName{
		ImplType:       implType,
		Module:         module,
		Implementation: implementation,
		Version:        "default",
	}
}

func NewSceneImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeScene, module, implementation, version)
}

func NewSceneImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeScene, module, implementation)
}

func NewCoreImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeCore, module, implementation, version)
}

func NewCoreImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeCore, module, implementation)
}

func NewModuleImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeModule, module, implementation, version)
}

func NewModuleImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeModule, module, implementation)
}

func NewInfraImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeInfra, module, implementation, version)
}

func NewInfraImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeInfra, module, implementation)
}

// NewModuleImplName creates a new module implementation name.
// Deprecated: use NewModuleImplNameNoVer instead
func NewAppImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeApp, module, implementation, version)
}

// NewModuleImplName creates a new module implementation name.
// Deprecated: use NewModuleImplNameNoVer instead
func NewAppImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeApp, module, implementation)
}

// Deprecated: use NewModuleImplNameNoVer instead
func NewSrvImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeSrv, module, implementation, version)
}

// Deprecated: use NewModuleImplNameNoVer instead
func NewSrvImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeSrv, module, implementation)
}

// Deprecated: use NewModuleImplNameNoVer instead
func NewRepoImplName(module, implementation, version string) ImplName {
	return NewImplName(ImplTypeRepo, module, implementation, version)
}

// Deprecated: use NewModuleImplNameNoVer instead
func NewRepoImplNameNoVer(module, implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeRepo, module, implementation)
}
