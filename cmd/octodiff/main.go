package main

import (
	"os"

	"github.com/OctopusDeploy/go-octodiff/pkg/cmd/root"
)

func main() {

	cmd := root.NewCmdRoot()
	cmd.AddCommand(NewCmdVersion())

	if err := cmd.Execute(); err != nil {
		cmd.PrintErr(err)
		cmd.Println()

		os.Exit(1)
	}
}
