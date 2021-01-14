package cmd

import (
	"fmt"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the LaunchControlD daemon",
	Long:  ``,
	Run:   setup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func setup(cmd *cobra.Command, args []string) {
	fmt.Println("Setup LaunchControlD started")
	start := time.Now()
	lctrld.SetupWorkspace(settings)
	lctrld.InstallDockerMachine(settings)
	fmt.Println("Setup completed in ", time.Since(start))
}
