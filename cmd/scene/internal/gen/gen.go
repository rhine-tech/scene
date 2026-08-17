package gen

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/cmd/scene/internal/gen/arpc"
	"github.com/spf13/cobra"
)

var CmdGen = &cobra.Command{
	Use:     "gen [gentype]",
	Short:   "code generation utility.",
	Version: scene.Version,
}

func init() {
	CmdGen.AddCommand(arpc.ARpcImplGen)
}
