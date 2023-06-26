package cmd

import (
	"context"
	"fmt"

	cmdc "github.com/jaimi-io/clobvm/cmd/clob-cli/consts"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/utils"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use: "balance",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, key, _, _, cli, err := defaultActor()
		if err != nil {
			return err
		}

		addr := key.PublicKey()

		if cmdc.GetAddress {
			addr, err = promptAddress("address")
			if err != nil {
				return err
			}
		}

		tokenID, err := promptToken("")
		if err != nil {
			return err
		}
		bal, err := cli.Balance(ctx, crypto.Address("clob", addr), tokenID)
		if err != nil {
			return err
		}
		format := "balance: %." + fmt.Sprint(consts.BalanceDecimals) + "f\n"
		fmt.Printf(format, bal)
		return nil
	},
}

var midPriceCmd = &cobra.Command{
	Use: "mid-price",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, _, _, cli, err := defaultActor()
		if err != nil {
			return err
		}

		baseTokenID, quoteTokenID := getTokens()
		if cmdc.GetPair {
			baseTokenID, err = promptToken("base")
			if err != nil {
				return err
			}

			quoteTokenID, err = promptToken("quote")
			if err != nil {
				return err
			}
		}

		pair := orderbook.Pair{BaseTokenID: baseTokenID, QuoteTokenID: quoteTokenID}
		
		midPrice, err := cli.MidPrice(ctx, pair)
		if err != nil {
			return err
		}
		fmt.Printf("mid price: %f\n", midPrice)
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
		if cmdc.GetPair {
			baseTokenID, err = promptToken("base")
			if err != nil {
				return err
			}

			quoteTokenID, err = promptToken("quote")
			if err != nil {
				return err
			}
		}


		numPriceLevels, err := promptInt("num price levels")
		if err != nil {
			return err
		}

		pair := orderbook.Pair{BaseTokenID: baseTokenID, QuoteTokenID: quoteTokenID}
		
		buySide, sellSide, err := cli.AllOrders(ctx, pair, numPriceLevels)
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

		if cmdc.GetAddress {
			addr, err = promptAddress("address")
			if err != nil {
				return err
			}
		}

		tokenID, err := promptToken("")
		if err != nil {
			return err
		}

		blockHeight, err := promptInt("block height")


		bal, resBlockHeight, err := cli.PendingFunds(ctx, addr, tokenID, uint64(blockHeight))
		if err != nil {
			return err
		}
		format := "pending balance: %." + fmt.Sprint(consts.BalanceDecimals) + "f\n"
		fmt.Printf(format, bal)
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
		if cmdc.GetPair {
			baseTokenID, err = promptToken("base")
			if err != nil {
				return err
			}

			quoteTokenID, err = promptToken("quote")
			if err != nil {
				return err
			}
		}

		pair := orderbook.Pair{BaseTokenID: baseTokenID, QuoteTokenID: quoteTokenID}

		numPriceLevels, err := promptInt("num price levels")
		if err != nil {
			return err
		}
		
		volumes, err := cli.Volumes(ctx, pair, numPriceLevels)
		if err != nil {
			return err
		}
		utils.Outf(volumes)
		return nil
	},
}