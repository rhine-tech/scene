package arpc

import (
	"bytes"
	"fmt"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/cmd/scene/internal/gen/astparser"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var ARpcImplGen = &cobra.Command{
	Use:     "arpc",
	Short:   "Generate ARPC implementation stubs for interfaces",
	Version: scene.Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args)
	},
}

var (
	generatePath string
	packageName  string
	goFile       string
)

func init() {
	ARpcImplGen.Flags().StringVarP(&generatePath, "output", "o", "./gen/arpcimpl", "Output path for generated files")
	ARpcImplGen.Flags().StringVarP(&packageName, "package", "p", "arpcimpl", "Package name for generated files")
	ARpcImplGen.Flags().StringVarP(&goFile, "gofile", "f", os.Getenv("GOFILE"), "go file")
}

// Function to generate code from template and parsed interfaces
func generateCode(iface astparser.Interface, pkgInterfaces *astparser.PackageInterfaces) (string, error) {
	tmpl, err := template.New("arpc").Funcs(template.FuncMap{
		"UpperFirst": func(val string) string {
			return strings.ToUpper(val[:1]) + val[1:]
		},
	}).Parse("package " + packageName + serviceTemplate)
	if err != nil {
		return "", err
	}
	resolvedImports := make([]string, 0)
	for _, pkgName := range iface.RequiredPackage {
		resolvedImports = append(resolvedImports, pkgInterfaces.ResolveImport(pkgName))
	}
	iface.RequiredPackage = resolvedImports
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, &iface)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func run(targetInterfaces []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	pkgInterfaces, err := astparser.ParseInterfaceFromFile(filepath.Join(cwd, goFile))
	if err != nil {
		return err
	}

	for _, iface := range pkgInterfaces.Interfaces {
		if !(len(targetInterfaces) == 0 || slices.Contains(targetInterfaces, iface.InterfaceName)) {
			continue
		}
		code, err := generateCode(iface, pkgInterfaces)
		if err != nil {
			return err
		}

		// Create output directory if it doesn't exist
		err = os.MkdirAll(generatePath, 0755)
		if err != nil {
			return err
		}

		// Generate file name based on the interface name
		fileName := fmt.Sprintf("%s/%s.scene.arpc.gen.go", generatePath, iface.InterfaceName)
		err = os.WriteFile(fileName, []byte(code), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
