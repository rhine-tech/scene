package astparser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rhine-tech/scene/utils/must"
	"github.com/rogpeppe/go-internal/modfile"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"
)

type Method struct {
	Name    string
	Args    []TypedVar
	Returns []TypedVar
}

type TypedVar struct {
	Name string
	Type string
}

type Interface struct {
	PackageName     string
	InterfaceName   string
	Methods         []Method
	RequiredPackage []string
}

func (i *Interface) addRequiredPkg(name string) {
	if slices.Contains(i.RequiredPackage, name) {
		return
	}
	i.RequiredPackage = append(i.RequiredPackage, name)
}

type PackageInterfaces struct {
	PackageName string
	Interfaces  []Interface
	Imports     map[string]string
}

func (i *PackageInterfaces) ResolveImport(name string) string {
	find, ok := i.Imports[name]
	if ok {
		return find
	}
	return "notfound"
}

func (i *PackageInterfaces) addImport(name string, path string) {
	_, ok := i.Imports[name]
	if ok {
		panic("import name " + name + " already exists")
	}
	i.Imports[name] = path
}

// findModuleRoot find go.mod file from current pwd. and its path to go mod
// adapt from go source code src/cmd/go/internal/modload/init.go
func findModuleRoot(dir string) (roots string, paths []string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)

	paths = make([]string, 0)

	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir, paths
		}
		d := filepath.Dir(dir)
		paths = append(paths, filepath.Base(dir))
		if d == dir {
			break
		}
		dir = d
	}
	return "", paths
}

func exprString(expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return buf.String()
}

func typeString(expr ast.Expr, currentPackage string) (string, map[string]string) {
	imports := make(map[string]string)
	switch t := expr.(type) {
	case *ast.Ident:
		if isPrimitiveType(t.Name) {
			return t.Name, imports
		}
		return fmt.Sprintf("%s.%s", currentPackage, t.Name), imports
	case *ast.SelectorExpr:
		packageName := t.X.(*ast.Ident).Name
		typeName := t.Sel.Name
		imports[packageName] = packageName
		return fmt.Sprintf("%s.%s", packageName, typeName), imports
		//return exprString(t)
	case *ast.StarExpr:
		elemType, elemImports := typeString(t.X, currentPackage)
		return "*" + elemType, elemImports
	case *ast.ArrayType:
		elemType, elemImports := typeString(t.Elt, currentPackage)
		return "[]" + elemType, elemImports
	case *ast.MapType:
		keyType, keyImports := typeString(t.Key, currentPackage)
		valueType, valueImports := typeString(t.Value, currentPackage)
		for k, v := range valueImports {
			keyImports[k] = v
		}
		return fmt.Sprintf("map[%s]%s", keyType, valueType), keyImports
		//return fmt.Sprintf("map[%s]%s", typeString(t.Key, currentPackage), typeString(t.Value, currentPackage)), imports
	case *ast.IndexExpr:
		// Handle single generic type
		xType, xImports := typeString(t.X, currentPackage)
		indexType, indexImports := typeString(t.Index, currentPackage)
		for k, v := range indexImports {
			xImports[k] = v
		}
		return fmt.Sprintf("%s[%s]", xType, indexType), xImports
	case *ast.IndexListExpr:
		// Handle multiple generic parameters
		xType, xImports := typeString(t.X, currentPackage)
		indexTypes := make([]string, len(t.Indices))
		for i, index := range t.Indices {
			indexType, indexImports := typeString(index, currentPackage)
			for k, v := range indexImports {
				imports[k] = v
			}
			indexTypes[i] = indexType
		}
		for k, v := range xImports {
			imports[k] = v
		}
		return fmt.Sprintf("%s[%s]", xType, strings.Join(indexTypes, ", ")), imports
	default:
		return exprString(t), imports
	}
}

func ParseInterfaceFromFile(path string) (*PackageInterfaces, error) {
	fset := token.NewFileSet()
	// 这里取绝对路径，方便打印出来的语法树可以转跳到编辑器
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	modRoot, importPaths := findModuleRoot(filepath.Dir(path))
	if modRoot == "" {
		return nil, errors.New("unable to find go.mod file")
	}
	fileNode, err := parser.ParseFile(fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	result := &PackageInterfaces{
		PackageName: fileNode.Name.Name,
		Interfaces:  make([]Interface, 0),
		Imports:     make(map[string]string),
	}
	modFile, err := modfile.Parse("go.mod", must.Must(os.ReadFile(filepath.Join(modRoot, "go.mod"))), nil)
	if err != nil {
		return nil, err
	}
	importPaths = append(importPaths, modFile.Module.Mod.Path)
	slices.Reverse(importPaths)
	result.Imports[result.PackageName] = strings.Join(importPaths, "/")
	for _, imports := range fileNode.Imports {
		var pkgPath, pkgName string
		pkgPath = imports.Path.Value
		pkgPath = pkgPath[1 : len(pkgPath)-1]
		if imports.Name != nil {
			pkgName = imports.Name.Name
		} else {
			pkgName = pkgPath[strings.LastIndex(pkgPath, "/")+1:]
		}
		result.addImport(pkgName, pkgPath)
	}
	for _, f := range fileNode.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range g.Specs {
			t, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			ifaceType, ok := t.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			iface := Interface{
				InterfaceName:   t.Name.Name,
				PackageName:     result.PackageName,
				Methods:         make([]Method, 0),
				RequiredPackage: make([]string, 0),
			}
			iface.addRequiredPkg(result.PackageName)
			for _, method := range ifaceType.Methods.List {
				// get all methods
				funcType, ok := method.Type.(*ast.FuncType)
				if !ok || len(method.Names) == 0 {
					continue
				}
				method := Method{
					Name:    method.Names[0].Name,
					Args:    make([]TypedVar, 0),
					Returns: make([]TypedVar, 0),
				}
				for _, param := range funcType.Params.List {
					argType, needed := typeString(param.Type, result.PackageName)
					for _, name := range param.Names {
						method.Args = append(method.Args, TypedVar{
							Name: name.Name,
							Type: argType,
						})
					}
					for pkg, _ := range needed {
						iface.addRequiredPkg(pkg)
					}
				}
				for _, param := range funcType.Results.List {
					argType, needed := typeString(param.Type, result.PackageName)
					for pkg, _ := range needed {
						iface.addRequiredPkg(pkg)
					}
					if param.Names == nil {
						method.Returns = append(method.Returns, TypedVar{
							Name: "",
							Type: argType,
						})
						continue
					}
					for _, name := range param.Names {
						method.Returns = append(method.Returns, TypedVar{
							Name: name.Name,
							Type: argType,
						})
					}
				}
				iface.Methods = append(iface.Methods, method)
			}
			result.Interfaces = append(result.Interfaces, iface)
		}
	}
	return result, nil
}
