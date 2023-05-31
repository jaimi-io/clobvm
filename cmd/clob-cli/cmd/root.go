package cmd

import (
	"github.com/jaimi-io/clobvm/cmd/clob-cli/consts"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/spf13/cobra"
)

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
		balanceCmd,
		allOrdersCmd,
	)

	rootCmd.PersistentPreRunE = func(*cobra.Command, []string) error {
		var err error
		consts.PrivKey, err = crypto.LoadKey("/home/jaimip/clobvm/.key.pk")
		if err != nil {
			return err
		}
		consts.ChainID, err = promptChainID()
		if err != nil {
			return err
		}
		consts.URI, err = promptURI()
		return err
	}

	actionCmd.AddCommand(
		transferCmd,
		addOrderCmd,
	)
}

func Execute() error {
	return rootCmd.Execute()
}