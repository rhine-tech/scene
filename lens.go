package scene

import (
	"fmt"
	"strings"
)

// Lens is an alias for Module

// Lens Definition

type ModuleName string

func (l ModuleName) String() string {
	return string(l)
}

func (l ModuleName) TableName(table string) string {
	return string(l) + "_" + table
}

func (l ModuleName) ImplName(iface, implementation string) ImplName {
	return NewImplName(ImplTypeModule, string(l), iface, implementation)
}

func (l ModuleName) ImplNameNoVer(implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeModule, string(l), implementation)
}

type InfraName string

func (i InfraName) String() string {
	return string(i)
}

func (i InfraName) ImplName(implementation, version string) ImplName {
	return NewImplName(ImplTypeInfra, string(i), implementation, version)
}

func (i InfraName) ImplNameNoVer(implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeInfra, string(i), implementation)
}

type CompositionName string

func (c CompositionName) String() string {
	return string(c)
}

func (c CompositionName) ImplName(implementation, version string) ImplName {
	return NewImplName(ImplTypeComp, string(c), implementation, version)
}

func (c CompositionName) ImplNameNoVer(implementation string) ImplName {
	return NewImplNameNoVer(ImplTypeComp, string(c), implementation)
}

/*
ImplName is the name of an implementation.
It is used to identify an implementation of a repository, service or application.
*/

type ImplType string

const (
	ImplTypeCore   = ImplType("core")      // core type
	ImplTypeScene  = ImplType("scene")     // scenario type
	ImplTypeInfra  = ImplType("infra")     // infrastructure type
	ImplTypeComp   = ImplType("composite") // composition type
	ImplTypeModule = ImplType("module")    // module type
)

type Named interface {
	ImplName() ImplName
}

type ImplName struct {
	ImplType       ImplType
	Module         string
	Interface      string
	Implementation string
}

// String returns a string representation of the implementation name.
// String is a pretty representation of the implementation name.
// If you want to use the implementation name as an identifier, use Identifier().
func (i ImplName) String() string {
	if i.Implementation == "" {
		return fmt.Sprintf("(%s)%s:%s", string(i.ImplType), i.Module, i.Interface)
	}
	return fmt.Sprintf("(%s)%s:%s:%s", string(i.ImplType), i.Module, i.Interface, i.Implementation)
}

func (i ImplName) EndpointName() string {
	return i.Module + "/" + i.Interface
}

// Identifier returns a string identifier of the implementation name.
func (i ImplName) Identifier() string {
	return fmt.Sprintf("(%s)%s:%s:%s", string(i.ImplType), i.Module, i.Interface, i.Implementation)
}

// ExportName return interface name with capitalized module name
func (i ImplName) ExportName() string {
	return fmt.Sprintf(strings.ToUpper(i.Module[:1]) + i.Module[1:] + "." + i.Interface)
}

func NewImplName(implType ImplType, module, implementation, version string) ImplName {
	return ImplName{
		ImplType:       implType,
		Module:         module,
		Interface:      implementation,
		Implementation: version,
	}
}

func NewImplNameNoVer(implType ImplType, module, iface string) ImplName {
	return ImplName{
		ImplType:       implType,
		Module:         module,
		Interface:      iface,
		Implementation: "default",
	}
}

func NewSceneImplName(module, iface, version string) ImplName {
	return NewImplName(ImplTypeScene, module, iface, version)
}

func NewSceneImplNameNoVer(module, iface string) ImplName {
	return NewImplNameNoVer(ImplTypeScene, module, iface)
}

func NewCoreImplName(module, iface, version string) ImplName {
	return NewImplName(ImplTypeCore, module, iface, version)
}

func NewCoreImplNameNoVer(module, iface string) ImplName {
	return NewImplNameNoVer(ImplTypeCore, module, iface)
}

func NewModuleImplName(module, iface, implementation string) ImplName {
	return NewImplName(ImplTypeModule, module, iface, implementation)
}

func NewModuleImplNameNoVer(module, iface string) ImplName {
	return NewImplNameNoVer(ImplTypeModule, module, iface)
}

func NewInfraImplName(module, iface, implementation string) ImplName {
	return NewImplName(ImplTypeInfra, module, iface, implementation)
}

func NewInfraImplNameNoVer(module, iface string) ImplName {
	return NewImplNameNoVer(ImplTypeInfra, module, iface)
}
