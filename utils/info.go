package utils

import (
	"fmt"
	"github.com/aynakeya/scene"
)

func FormatContainerInfo(idx int, container scene.ApplicationContainer) string {
	appNames := container.ListAppNames()
	nameStr := ""
	for i, name := range appNames {
		nameStr += fmt.Sprintf("%d. %s\n", i, name)
	}
	return fmt.Sprintf("#%d %s: %d App loaded\n-------\n%s\n", idx+1, container.Name(), len(appNames), nameStr)
}
