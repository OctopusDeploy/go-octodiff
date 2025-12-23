package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "0.0.0-dev"
)

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "version",
		Long: "Reports the version of octodiff",
		RunE: func(c *cobra.Command, args []string) error {
			// pick up positional arguments if not explicitly specified using --basis-file and --signature-file
			fmt.Printf("App Version: %s\n", version)

			if buildInfo, ok := debug.ReadBuildInfo(); ok {
				for _, setting := range buildInfo.Settings {
					if setting.Key == "vcs.revision" {
						fmt.Printf("Commit Hash: %s\n", setting.Value)
					}
					if setting.Key == "vcs.time" {
						fmt.Printf("Build Time: %s\n", setting.Value)
					}
				}
				fmt.Printf("Go Version: %s\n", buildInfo.GoVersion)
			}

			return nil
		},
	}

	return cmd
}
