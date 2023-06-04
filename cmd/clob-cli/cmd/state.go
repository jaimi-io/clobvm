package cmd

import (
	"context"
	"fmt"

	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use: "balance",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, _, _, cli, err := defaultActor()
		if err != nil {
			return err
		}
		addr, err := promptAddress("address")
		if err != nil {
			return err
		}
		tokenID, err := promptToken("")
		if err != nil {
			return err
		}
		bal, err := cli.Balance(ctx, crypto.Address("clob", addr), tokenID)
		if err != nil {
			return err
		}
		fmt.Printf("balance: %d\n", bal)
		return nil
	},
}

var allOrdersCmd = &cobra.Command{
	Use: "orders",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, _, _, cli, err := defaultActor()
		if err != nil {
			return err
		}

		baseTokenID, quoteTokenID := getTokens()
		pair := orderbook.Pair{BaseTokenID: baseTokenID, QuoteTokenID: quoteTokenID}
		
		buySide, sellSide, err := cli.AllOrders(ctx, pair)
		if err != nil {
			return err
		}
		fmt.Printf("sell side: %s\n", sellSide)
		fmt.Printf("buy side: %s\n", buySide)
		return nil
	},
}

var pendingFundsCmd = &cobra.Command{
	Use: "pending",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, key, _, _, cli, err := defaultActor()
		if err != nil {
			return err
		}
		addr := key.PublicKey()

		tokenID, err := promptToken("")
		if err != nil {
			return err
		}

		blockHeight, err := promptInt("block height")


		bal, resBlockHeight, err := cli.PendingFunds(ctx, addr, tokenID, uint64(blockHeight))
		if err != nil {
			return err
		}
		fmt.Printf("pending balance: %d\n", bal)
		fmt.Printf("at block height: %d\n", resBlockHeight)
		return nil
	},
}

var volumesCmd = &cobra.Command{
	Use: "volumes",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, _, _, cli, err := defaultActor()
		if err != nil {
			return err
		}

		baseTokenID, quoteTokenID := getTokens()
		pair := orderbook.Pair{BaseTokenID: baseTokenID, QuoteTokenID: quoteTokenID}
		
		volumes, err := cli.Volumes(ctx, pair)
		if err != nil {
			return err
		}
		fmt.Printf(volumes)
		return nil
	},
}