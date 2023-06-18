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

	rootCmd.PersistentPreRunE = func(*cobra.Command, []string) error {
		var err error
		consts.PrivKey, err = crypto.LoadKey("/home/jaimip/clobvm/.key.pk")
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
	)

	prometheusCmd.AddCommand(
		generatePrometheusCmd,
	)

	spamCmd.AddCommand(
		transferSpamCmd,
		orderSpamCmd,
		simulateOrderCmd,
	)
}

func Execute() error {
	return rootCmd.Execute()
}