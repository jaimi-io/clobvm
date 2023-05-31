package cmd

import (
	"context"
	"fmt"

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
		tokenID, err := promptToken()
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
		buySide, sellSide, err := cli.AllOrders(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("buy side: %s\n", buySide)
		fmt.Printf("sell side: %s\n", sellSide)
		return nil
	},
}