package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Get help about any command",
	Long: `Help provides help for any command in the application.
	
Simply type 'openshift-metadata-manager help [command]' for full details.`,
	Run: func(cmd *cobra.Command, args []string) {
		target, _, _ := cmd.Root().Find(args)
		if target != nil {
			_ = target.Help()
		} else {
			fmt.Printf("Unknown help topic: %q\n", args)
		}
	},
}

func init() {
	RootCmd.AddCommand(helpCmd)
}
