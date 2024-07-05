package scene

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImplName_ExportName(t *testing.T) {
	var name ImplName = NewModuleImplNameNoVer("hello", "World")
	require.Equal(t, "Hello.World", name.ExportName())

}
