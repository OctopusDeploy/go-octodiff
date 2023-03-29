package root

import (
	"github.com/OctopusDeploy/go-octodiff/pkg/cmd/delta"
	"github.com/OctopusDeploy/go-octodiff/pkg/cmd/explaindelta"
	"github.com/OctopusDeploy/go-octodiff/pkg/cmd/patch"
	"github.com/OctopusDeploy/go-octodiff/pkg/cmd/signature"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use: "octodiff <command>",
	}

	cmd.AddCommand(signature.NewCmdSignature())
	cmd.AddCommand(delta.NewCmdDelta())
	cmd.AddCommand(patch.NewCmdPatch())
	cmd.AddCommand(explaindelta.NewCmdExplainDelta())

	return cmd
}
