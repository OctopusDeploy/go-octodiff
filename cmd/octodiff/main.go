package main

import (
	"github.com/OctopusDeploy/go-octodiff/pkg/cmd/root"
	"os"
)

func main() {
	cmd := root.NewCmdRoot()

	if err := cmd.Execute(); err != nil {
		cmd.PrintErr(err)
		cmd.Println()

		os.Exit(1)
	}
}
