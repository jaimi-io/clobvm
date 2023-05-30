package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:        "clob-cli",
		Short:      "ClobVM CLI",
		SuggestFor: []string{"clob-cli", "clobcli"},
	}
)

func init() {
	cobra.EnablePrefixMatching = true
	rootCmd.AddCommand(
		actionCmd,
	)

	actionCmd.AddCommand(
		transferCmd,
	)
}

func Execute() error {
	return rootCmd.Execute()
}