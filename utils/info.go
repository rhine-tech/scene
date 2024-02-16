package utils

import (
	"fmt"
	"github.com/rhine-tech/scene"
	"strings"
)

func FormatContainerInfo(idx int, container scene.ApplicationContainer) string {
	appNames := container.ListAppNames()
	nameStr := ""
	padding := strings.Repeat(" ", 8)
	sep := strings.Repeat("-", 64)
	for i, name := range appNames {
		nameStr += fmt.Sprintf("%s%d. %s\n", padding, i+1, name)
	}
	return fmt.Sprintf("#%d %s: %d App loaded\n%s\n%s\n", idx+1, container.Name(), len(appNames), sep, nameStr)
}
