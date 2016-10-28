package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is root cmd of dockyard.
var RootCmd = &cobra.Command{
	Use:   "wharf",
	Short: "Wharf Is Agile Project Management of ContainerOps",
	Long:  `Wharf is the agile project management part of containerops.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// init()
func init() {

}
