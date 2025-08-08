package gen

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/cmd/scene/internal/gen/arpc"
	"github.com/rhine-tech/scene/cmd/scene/internal/gen/rpc"
	"github.com/spf13/cobra"
)

var CmdGen = &cobra.Command{
	Use:     "gen [gentype]",
	Short:   "code generation utility.",
	Version: scene.Version,
}

func init() {
	CmdGen.AddCommand(rpc.RpcImplGen)
	CmdGen.AddCommand(arpc.ARpcImplGen)
}
