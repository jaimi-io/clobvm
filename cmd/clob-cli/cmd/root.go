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
		prometheusCmd,
		spamCmd,
		pendingFundsCmd,
		volumesCmd,
		midPriceCmd,
	)

	rootCmd.PersistentFlags().BoolVar(&consts.GetPair, "get-pair", false, "get pair from user input")
	rootCmd.PersistentFlags().BoolVar(&consts.GetAddress, "get-address", false, "get address from user input")
	rootCmd.PersistentFlags().StringVar(
		&consts.KeyPath,
		"pk",
		".key.pk",
		"path to private key file",
	)

	rootCmd.PersistentPreRunE = func(*cobra.Command, []string) error {
		var err error
		consts.PrivKey, err = crypto.LoadKey(consts.KeyPath)
		if err != nil {
			return err
		}
		consts.URI, consts.URIS, err = promptURIs()
		if err != nil {
			return err
		}
		consts.ChainID, err = promptChainID()
		return err
	}

	actionCmd.AddCommand(
		transferCmd,
		addOrderCmd,
		cancelOrderCmd,
		marketOrderCmd,
		cancelAllOrderCmd,
	)

	spamCmd.AddCommand(
		transferSpamCmd,
		orderMatchSpamCmd,
		orderDeterministicSpamCmd,
		simulateOrderCmd,
	)
}

func Execute() error {
	return rootCmd.Execute()
}