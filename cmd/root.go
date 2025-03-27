package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	kubeconfigPath string
	platform       string
	dryRun         bool
)

var RootCmd = &cobra.Command{
	Use:   "openshift-metadata-manager",
	Short: "OpenShift infrastructure resource tagger",
	Long: `A tool to manage tags on cloud infrastructure resources for OpenShift clusters.

Examples:
  # Sync tags for auto-detected platform
  openshift-metadata-manager sync
  
  # Sync tags for AWS with dry-run
  openshift-metadata-manager sync --platform aws --dry-run
  
  # Show version information
  openshift-metadata-manager version`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

// Add a version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("openshift-metadata-manager v0.1.0")
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", "", "Path to kubeconfig file")
	RootCmd.PersistentFlags().StringVarP(&platform, "platform", "p", "", "Override cloud platform (aws, azure, gcp)")
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
