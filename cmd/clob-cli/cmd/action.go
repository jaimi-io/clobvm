package cmd

import (
	"context"
	"errors"

	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/spf13/cobra"
)

var actionCmd = &cobra.Command{
	Use: "action",
	RunE: func(*cobra.Command, []string) error {
		return errors.New("subcommand not implemented")
	},
}

var transferCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, authFactory, cli, tcli, err := defaultActor()
		if err != nil {
			return err
		}
		tokenID, err := promptToken("")
		if err != nil {
			return err
		}

		// Select recipient
		recipient, err := promptAddress("recipient")
		if err != nil {
			return err
		}
		
		// Select amount
		amount, err := promptAmount("amount", consts.BalanceDecimals)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := promptContinue()
		if !cont || err != nil {
			return err
		}

		parser, err := tcli.Parser(ctx)
		if err != nil {
			return err
		}

		// Generate transaction
		submit, _, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
			To:    recipient,
			TokenID: tokenID,
			Amount: amount,
		}, authFactory)
		if err != nil {
			return err
		}
		if err := submit(ctx); err != nil {
			return err
		}

		return nil
	},
}

var addOrderCmd = &cobra.Command{
	Use: "add-limit-order",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, authFactory, cli, tcli, err := defaultActor()
		if err != nil {
			return err
		}
		baseTokenID, err := promptToken("base")
		if err != nil {
			return err
		}

		quoteTokenID, err := promptToken("quote")
		if err != nil {
			return err
		}
		
		quantity, err := promptAmount("quantity", consts.BalanceDecimals)
		if err != nil {
			return err
		}

		side, err := promptBool("side")
		if err != nil {
			return err
		}

		price, err := promptAmount("price", consts.PriceDecimals)
		if err != nil {
			return err
		}

		blockExpiryWindow, err := promptOptional("blockExpiryWindow")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := promptContinue()
		if !cont || err != nil {
			return err
		}

		parser, err := tcli.Parser(ctx)
		if err != nil {
			return err
		}

		// Generate transaction
		submit, _, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.AddOrder{
			Pair: orderbook.Pair{
				BaseTokenID: baseTokenID,
				QuoteTokenID: quoteTokenID,
			},
			Quantity: quantity,
			Price: price,
			Side: side,
			BlockExpiryWindow: uint64(blockExpiryWindow),
		}, authFactory)
		if err != nil {
			return err
		}
		if err := submit(ctx); err != nil {
			return err
		}
		return nil
	},
}

var marketOrderCmd = &cobra.Command{
	Use: "add-market-order",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, authFactory, cli, tcli, err := defaultActor()
		if err != nil {
			return err
		}
		baseTokenID, err := promptToken("base")
		if err != nil {
			return err
		}

		quoteTokenID, err := promptToken("quote")
		if err != nil {
			return err
		}
		
		quantity, err := promptAmount("quantity", consts.BalanceDecimals)
		if err != nil {
			return err
		}

		side, err := promptBool("side")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := promptContinue()
		if !cont || err != nil {
			return err
		}

		parser, err := tcli.Parser(ctx)
		if err != nil {
			return err
		}

		// Generate transaction
		submit, _, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.AddOrder{
			Pair: orderbook.Pair{
				BaseTokenID: baseTokenID,
				QuoteTokenID: quoteTokenID,
			},
			Quantity: quantity,
			Side: side,
		}, authFactory)
		if err != nil {
			return err
		}
		if err := submit(ctx); err != nil {
			return err
		}
		return nil
	},
}

var cancelOrderCmd = &cobra.Command{
	Use: "cancel-order",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, authFactory, cli, tcli, err := defaultActor()
		if err != nil {
			return err
		}
		baseTokenID, err := promptToken("base")
		if err != nil {
			return err
		}

		quoteTokenID, err := promptToken("quote")
		if err != nil {
			return err
		}
		
		orderID, err := promptID("orderID")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := promptContinue()
		if !cont || err != nil {
			return err
		}

		parser, err := tcli.Parser(ctx)
		if err != nil {
			return err
		}

		// Generate transaction
		submit, _, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.CancelOrder{
			Pair: orderbook.Pair{
				BaseTokenID: baseTokenID,
				QuoteTokenID: quoteTokenID,
			},
			OrderID: orderID,
		}, authFactory)
		if err != nil {
			return err
		}
		if err := submit(ctx); err != nil {
			return err
		}
		return nil
	},
}
