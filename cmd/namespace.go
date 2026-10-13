package cmd

import (
	"fmt"

	"github.com/janklabs/obscuro/internal/store"
	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:     "namespace",
	Aliases: []string{"namespaces"},
	Short:   "Manage secret namespaces",
	Long: `Namespaces group secrets within one repository vault and share its master password.
Select a namespace with --namespace/-n on secret commands. The default namespace
contains existing secrets. Named namespaces are created automatically by set or import.`,
}

var namespaceListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List namespaces without decrypting secrets",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !store.IsInitialized() {
			return fmt.Errorf("not initialized — run 'obscuro init' first")
		}
		names, err := store.ListNamespaces()
		if err != nil {
			return err
		}
		for _, name := range names {
			fmt.Fprintln(Stdout, name)
		}
		return nil
	},
}

func init() {
	namespaceCmd.AddCommand(namespaceListCmd)
	rootCmd.AddCommand(namespaceCmd)
}
