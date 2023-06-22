package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	amath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/cmd/clob-cli/consts"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/orderbook"
	trpc "github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/codec"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/pubsub"
	"github.com/jaimi-io/hypersdk/rpc"
	hutils "github.com/jaimi-io/hypersdk/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	feePerTx     = 1000
)

type txIssuer struct {
	c  *rpc.JSONRPCClient
	tc *trpc.JSONRPCClient
	d  *rpc.WebSocketClient

	l              sync.Mutex
	outstandingTxs int
}
type timeModifier struct {
	Timestamp int64
}

func (t *timeModifier) Base(b *chain.Base) {
	b.Timestamp = t.Timestamp
}

var balance = uint64(1_000_000_000 * utils.MinBalance())

var spamCmd = &cobra.Command{
	Use: "spam",
	RunE: func(*cobra.Command, []string) error {
		return errors.New("must specify a subcommand")
	},
}

func getRandomRecipient(self int, keys []crypto.PrivateKey) (crypto.PublicKey, error) {
	// Select item from array
	index := rand.Int() % len(keys)
	if index == self {
		index++
		if index == len(keys) {
			index = 0
		}
	}
	return keys[index].PublicKey(), nil
}

type BalanceUpdate struct {
	BaseTokenUser  crypto.PublicKey
	BaseBalance	   uint64
	QuoteTokenUser crypto.PublicKey
	QuoteBalance	 uint64
}

func getBalanceUpdate(output []byte) *BalanceUpdate {
	p := codec.NewReader(output, math.MaxInt)
	balUp := &BalanceUpdate{}
	p.UnpackPublicKey(true, &balUp.BaseTokenUser)
	balUp.BaseBalance = p.UnpackUint64(false)
	p.UnpackPublicKey(true, &balUp.QuoteTokenUser)
	balUp.QuoteBalance = p.UnpackUint64(false)
	return balUp
}

func getRandomIssuer(issuers []*txIssuer) *txIssuer {
	index := rand.Int() % len(issuers)
	return issuers[index]
}

var transferSpamCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		chainID, key, _, _, _, err := defaultActor()
		if err != nil {
			return err
		}

		uris := consts.URIS
		cli := rpc.NewJSONRPCClient(uris[0])
		tcli := trpc.NewRPCClient(uris[0], chainID, genesis.New())
		factory := auth.NewEIP712Factory(key)
		avaxID, _ := getTokens()


		// Distribute funds
		numAccounts, err := promptInt("number of accounts")
		if err != nil {
			return err
		}
		numTxsPerAccount, err := promptInt("number of transactions per account per second")
		if err != nil {
			return err
		}
		witholding := uint64(feePerTx * numAccounts)
		distAmount := (balance - witholding) / uint64(numAccounts)
		hutils.Outf(
			"{{yellow}}distributing funds to each account:{{/}} %s %s\n",
			distAmount,
			avaxID,
		)
		accounts := make([]crypto.PrivateKey, numAccounts)
		dcli, err := rpc.NewWebSocketClient(uris[0], 8_192, pubsub.MaxReadMessageSize)
		if err != nil {
			return err
		}
		funds := map[crypto.PublicKey]uint64{}
		parser, err := tcli.Parser(ctx)
		if err != nil {
			return err
		}
		var fundsL sync.Mutex
		for i := 0; i < numAccounts; i++ {
			// Create account
			pk, err := crypto.GeneratePrivateKey()
			if err != nil {
				return err
			}
			accounts[i] = pk

			// Send funds
			_, tx, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    pk.PublicKey(),
				TokenID: avaxID,
				Amount: distAmount,
			}, factory)
			if err != nil {
				return err
			}
			if err := dcli.RegisterTx(tx); err != nil {
				return err
			}

			// Ensure Snowman++ is activated
			if i < 10 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		for i := 0; i < numAccounts; i++ {
			_, dErr, result, err := dcli.ListenTx(ctx)
			if err != nil {
				return err
			}
			if dErr != nil {
				return dErr
			}
			if !result.Success {
				fmt.Println(string(result.Output))
				// Should never happen
				return errors.New("failed to return funds")
			}
			balUp := getBalanceUpdate(result.Output)
			funds[balUp.BaseTokenUser] = balUp.BaseBalance
			funds[balUp.QuoteTokenUser] = balUp.QuoteBalance
		}
		hutils.Outf("{{yellow}}distributed funds to %d accounts{{/}}\n", numAccounts)
		signals := make(chan os.Signal, 2)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		var (
			transferFee uint64
			wg          sync.WaitGroup

			l            sync.Mutex
			confirmedTxs uint64
			totalTxs     uint64
		)

		clients := make([]*txIssuer, len(uris))
		for i := 0; i < len(uris); i++ {
			cli := rpc.NewJSONRPCClient(uris[i])
			tcli := trpc.NewRPCClient(uris[i], chainID, genesis.New())
			dcli, err := rpc.NewWebSocketClient(uris[i], 128_000, pubsub.MaxReadMessageSize)
			if err != nil {
				return err
			}
			clients[i] = &txIssuer{c: cli, tc: tcli, d: dcli}
		}

		// confirm txs (track failure rate)
		cctx, cancel := context.WithCancel(ctx)
		defer cancel()
		var inflight atomic.Int64
		var sent atomic.Int64
		var exiting sync.Once
		for i := 0; i < len(clients); i++ {
			issuer := clients[i]
			wg.Add(1)
			go func() {
				for {
					_, dErr, result, err := issuer.d.ListenTx(context.TODO())
					if err != nil {
						return
					}
					inflight.Add(-1)
					issuer.l.Lock()
					issuer.outstandingTxs--
					issuer.l.Unlock()
					l.Lock()
					if result != nil {
						if result.Success {
							confirmedTxs++
							balUp := getBalanceUpdate(result.Output)
							fundsL.Lock()
							funds[balUp.BaseTokenUser] = balUp.BaseBalance
							funds[balUp.QuoteTokenUser] = balUp.QuoteBalance
							fundsL.Unlock()
						} else {
							hutils.Outf("{{orange}}on-chain tx failure:{{/}} %s %t\n", string(result.Output), result.Success)
						}
					} else {
						// We can't error match here because we receive it over the wire.
						if !strings.Contains(dErr.Error(), rpc.ErrExpired.Error()) {
							hutils.Outf("{{orange}}pre-execute tx failure:{{/}} %v\n", dErr)
						}
					}
					totalTxs++
					l.Unlock()
				}
			}()
			go func() {
				<-cctx.Done()
				for {
					issuer.l.Lock()
					outstanding := issuer.outstandingTxs
					issuer.l.Unlock()
					if outstanding == 0 {
						_ = issuer.d.Close()
						wg.Done()
						return
					}
					time.Sleep(500 * time.Millisecond)
				}
			}()
		}

		// log stats
		t := time.NewTicker(1 * time.Second) // ensure no duplicates created
		defer t.Stop()
		var psent int64
		go func() {
			for {
				select {
				case <-t.C:
					current := sent.Load()
					l.Lock()
					if totalTxs > 0 {
						hutils.Outf(
							"{{yellow}}txs seen:{{/}} %d {{yellow}}success rate:{{/}} %.2f%% {{yellow}}inflight:{{/}} %d {{yellow}}issued/s:{{/}} %d\n", //nolint:lll
							totalTxs,
							float64(confirmedTxs)/float64(totalTxs)*100,
							inflight.Load(),
							current-psent,
						)
					}
					l.Unlock()
					psent = current
				case <-cctx.Done():
					return
				}
			}
		}()

		// broadcast txs
		if err != nil {
			return err
		}
		g, gctx := errgroup.WithContext(ctx)
		for ri := 0; ri < numAccounts; ri++ {
			i := ri
			g.Go(func() error {
				t := time.NewTimer(0) // ensure no duplicates created
				defer t.Stop()

				issuer := getRandomIssuer(clients)
				factory := auth.NewEIP712Factory(accounts[i])
				for {
					select {
					case <-t.C:
						// Ensure we aren't too backlogged
						maxTxBacklog := 72_000
						if inflight.Load() > int64(maxTxBacklog) {
							t.Reset(1 * time.Second)
							continue
						}

						// Send transaction
						start := time.Now()
						selected := map[crypto.PublicKey]int{}
						for k := 0; k < numTxsPerAccount; k++ {
							recipient, err := getRandomRecipient(i, accounts)
							if err != nil {
								return err
							}
							v := selected[recipient] + 1
							selected[recipient] = v
							_, tx, fees, err := issuer.c.GenerateTransactionManual(parser, nil, &actions.Transfer{
								To:    recipient,
								TokenID: avaxID,
								Amount: uint64(v), // ensure txs are unique
							}, factory, 0)
							if err != nil {
								hutils.Outf("{{orange}}failed to generate:{{/}} %v\n", err)
								continue
							}
							transferFee = fees
							if err := issuer.d.RegisterTx(tx); err != nil {
								continue
							}
							balance -= (fees + uint64(v))
							issuer.l.Lock()
							issuer.outstandingTxs++
							issuer.l.Unlock()
							inflight.Add(1)
							sent.Add(1)
						}

						// Determine how long to sleep
						dur := time.Since(start)
						sleep := amath.Max(1000-dur.Milliseconds(), 0)
						t.Reset(time.Duration(sleep) * time.Millisecond)
					case <-gctx.Done():
						return gctx.Err()
					case <-cctx.Done():
						return nil
					case <-signals:
						exiting.Do(func() {
							hutils.Outf("{{yellow}}exiting broadcast loop{{/}}\n")
							cancel()
						})
						return nil
					}
				}
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}

		// Wait for all issuers to finish
		hutils.Outf("{{yellow}}waiting for issuers to return{{/}}\n")
		dctx, cancel := context.WithCancel(ctx)
		go func() {
			// Send a dummy transaction if shutdown is taking too long (listeners are
			// expired on accept if dropped)
			t := time.NewTicker(15 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					hutils.Outf("{{yellow}}remaining:{{/}} %d\n", inflight.Load())
					_ = submitDummy(dctx, cli, tcli, key.PublicKey(), factory)
				case <-dctx.Done():
					return
				}
			}
		}()
		wg.Wait()
		cancel()

		// Return funds
		hutils.Outf("{{yellow}}returning funds to %s{{/}}\n", crypto.Address("clob", key.PublicKey()))
		var (
			returnedBalance uint64
			returnsSent     int
		)
		for i := 0; i < numAccounts; i++ {
			balance := funds[accounts[i].PublicKey()]
			if transferFee > balance {
				continue
			}
			returnsSent++
			// Send funds
			returnAmt := balance - transferFee
			_, tx, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    key.PublicKey(),
				TokenID: avaxID,
				Amount: returnAmt,
			}, auth.NewEIP712Factory(accounts[i]))
			if err != nil {
				return err
			}
			if err := dcli.RegisterTx(tx); err != nil {
				return err
			}
			returnedBalance += returnAmt

			// Ensure Snowman++ is activated
			if i < 10 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		for i := 0; i < returnsSent; i++ {
			_, dErr, result, err := dcli.ListenTx(ctx)
			if err != nil {
				return err
			}
			if dErr != nil {
				return dErr
			}
			if !result.Success {
				// Should never happen
				return errors.New("failed to return funds")
			}
		}
		hutils.Outf(
			"{{yellow}}returned funds:{{/}} %s %s\n",
			returnedBalance,
			avaxID,
		)
		return nil
	},
}

var orderSpamCmd = &cobra.Command{
	Use: "order",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		chainID, key, _, _, _, err := defaultActor()
		if err != nil {
			return err
		}

		uris := consts.URIS
		cli := rpc.NewJSONRPCClient(uris[0])
		tcli := trpc.NewRPCClient(uris[0], chainID, genesis.New())
	
		factory := auth.NewEIP712Factory(key)
		avaxID, usdcID := getTokens()
		pair := orderbook.Pair{
			BaseTokenID:  avaxID,
			QuoteTokenID: usdcID,
		}
		buyPrices := []uint64{
			1,
			2,
			3,
			4,
			5,
		}
		sellPrices := []uint64{
			3,
			4,
			5,
			6,
			7,
		}

		// Distribute funds
		numAccounts, err := promptInt("number of accounts")
		if err != nil {
			return err
		}
		numTxsPerAccount, err := promptInt("number of transactions per account per second")
		if err != nil {
			return err
		}
		witholding := uint64(feePerTx * numAccounts)
		distAmount := (balance - witholding) / uint64(numAccounts)
		hutils.Outf(
			"{{yellow}}distributing funds to each account:{{/}} %s %s\n",
			distAmount,
			avaxID,
		)
		accounts := make([]crypto.PrivateKey, numAccounts)
		dcli, err := rpc.NewWebSocketClient(uris[0], 8_192, pubsub.MaxReadMessageSize)
		if err != nil {
			return err
		}
		funds := map[crypto.PublicKey]map[ids.ID]uint64{}
		parser, err := tcli.Parser(ctx)
		if err != nil {
			return err
		}
		clients := make([]*txIssuer, len(uris))
		for i := 0; i < len(uris); i++ {
			cli := rpc.NewJSONRPCClient(uris[i])
			tcli := trpc.NewRPCClient(uris[i], chainID, genesis.New())
			dcli, err := rpc.NewWebSocketClient(uris[i], 128_000, pubsub.MaxReadMessageSize)
			if err != nil {
				return err
			}
			clients[i] = &txIssuer{c: cli, tc: tcli, d: dcli}
		}
		funds[key.PublicKey()] = make(map[ids.ID]uint64)
		var fundsL sync.Mutex
		for i := 0; i < numAccounts; i++ {
			// Create account
			pk, err := crypto.GeneratePrivateKey()
			if err != nil {
				return err
			}
			accounts[i] = pk
			funds[pk.PublicKey()] = make(map[ids.ID]uint64)

			// Send funds
			_, tx, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    pk.PublicKey(),
				TokenID: avaxID,
				Amount: distAmount,
			}, factory)
			if err != nil {
				return err
			}
			if err := dcli.RegisterTx(tx); err != nil {
				return err
			}

			// Ensure Snowman++ is activated
			if i < 10 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		for i := 0; i < numAccounts; i++ {
			_, dErr, result, err := dcli.ListenTx(ctx)
			if err != nil {
				return err
			}
			if dErr != nil {
				return dErr
			}
			if !result.Success {
				// Should never happen
				return errors.New("failed to return funds")
			}
			balUp := getBalanceUpdate(result.Output)
			funds[balUp.BaseTokenUser][avaxID] = balUp.BaseBalance
			funds[balUp.QuoteTokenUser][avaxID] = balUp.QuoteBalance
		}
		hutils.Outf("{{yellow}}distributed avax funds to %d accounts{{/}}\n", numAccounts)

		for i := 0; i < numAccounts; i++ {
			// Create account
			pk := accounts[i]

			_, tx, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    pk.PublicKey(),
				TokenID: usdcID,
				Amount: distAmount,
			}, factory)
			if err != nil {
				return err
			}
			if err := dcli.RegisterTx(tx); err != nil {
				return err
			}

			// Ensure Snowman++ is activated
			if i < 10 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		for i := 0; i < numAccounts; i++ {
			_, dErr, result, err := dcli.ListenTx(ctx)
			if err != nil {
				return err
			}
			if dErr != nil {
				return dErr
			}
			if !result.Success {
				// Should never happen
				return errors.New("failed to return funds")
			}
			balUp := getBalanceUpdate(result.Output)
			funds[balUp.BaseTokenUser][usdcID] = balUp.BaseBalance
			funds[balUp.QuoteTokenUser][usdcID] = balUp.QuoteBalance
		}
		hutils.Outf("{{yellow}}distributed usdc funds to %d accounts{{/}}\n", numAccounts)
		signals := make(chan os.Signal, 2)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		var (
			transferFee uint64
			wg          sync.WaitGroup

			l            sync.Mutex
			confirmedTxs uint64
			totalTxs     uint64
		)

		// confirm txs (track failure rate)
		cctx, cancel := context.WithCancel(ctx)
		defer cancel()
		var inflight atomic.Int64
		var sent atomic.Int64
		var exiting sync.Once
		for i := 0; i < len(clients); i++ {
			issuer := clients[i]
			wg.Add(1)
			go func() {
				for {
					_, dErr, result, err := issuer.d.ListenTx(context.TODO())
					if err != nil {
						fmt.Println(err)
						return
					}
					inflight.Add(-1)
					issuer.l.Lock()
					issuer.outstandingTxs--
					issuer.l.Unlock()
					l.Lock()
					if result != nil {
						if result.Success {
							confirmedTxs++
							balUp := getBalanceUpdate(result.Output)
							fundsL.Lock()
							if balUp.BaseBalance > 0 {
								funds[balUp.BaseTokenUser][avaxID] = balUp.BaseBalance
							}
							if balUp.QuoteBalance > 0 {
								funds[balUp.QuoteTokenUser][usdcID] = balUp.QuoteBalance
							}
							fundsL.Unlock()
						} else {
							hutils.Outf("{{orange}}on-chain tx failure:{{/}} %s %t\n", string(result.Output), result.Success)
						}
					} else {
						// We can't error match here because we receive it over the wire.
						if !strings.Contains(dErr.Error(), rpc.ErrExpired.Error()) {
							hutils.Outf("{{orange}}pre-execute tx failure:{{/}} %v\n", dErr)
						}
					}
					totalTxs++
					l.Unlock()
				}
			}()
			go func() {
				<-cctx.Done()
				for {
					issuer.l.Lock()
					outstanding := issuer.outstandingTxs
					issuer.l.Unlock()
					if outstanding == 0 {
						_ = issuer.d.Close()
						wg.Done()
						return
					}
					time.Sleep(500 * time.Millisecond)
				}
			}()
		}

		// log stats
		t := time.NewTicker(1 * time.Second) // ensure no duplicates created
		defer t.Stop()
		var psent int64
		go func() {
			for {
				select {
				case <-t.C:
					current := sent.Load()
					l.Lock()
					if totalTxs > 0 {
						hutils.Outf(
							"{{yellow}}txs seen:{{/}} %d {{yellow}}success rate:{{/}} %.2f%% {{yellow}}inflight:{{/}} %d {{yellow}}issued/s:{{/}} %d\n", //nolint:lll
							totalTxs,
							float64(confirmedTxs)/float64(totalTxs)*100,
							inflight.Load(),
							current-psent,
						)
					}
					l.Unlock()
					psent = current
				case <-cctx.Done():
					return
				}
			}
		}()

		// broadcast txs
		if err != nil {
			return err
		}
		g, gctx := errgroup.WithContext(ctx)
		for ri := 0; ri < numAccounts; ri++ {
			i := ri
			g.Go(func() error {
				t := time.NewTimer(0) // ensure no duplicates created
				defer t.Stop()

				issuer := getRandomIssuer(clients)
				factory := auth.NewEIP712Factory(accounts[i])
				ut := time.Now().Unix()
				for {
					select {
					case <-t.C:
						// Ensure we aren't too backlogged
						maxTxBacklog := 72_000
						if inflight.Load() > int64(maxTxBacklog) {
							t.Reset(1 * time.Second)
							continue
						}

						nextTime := time.Now().Unix()
						if nextTime <= ut {
							nextTime = ut + 1
						}
						ut = nextTime

						// Send transaction
						start := time.Now()
						selected := map[crypto.PublicKey]uint64{}
						tm := &timeModifier{nextTime + parser.Rules(nextTime).GetValidityWindow() - 3}
						for a:=0; a<numAccounts; a++ {
							selected[accounts[a].PublicKey()] = utils.MinQuantity()
						}
						for k := 0; k < numTxsPerAccount; k++ {
							v := selected[accounts[i].PublicKey()] + utils.MinQuantity()
							selected[accounts[i].PublicKey()] = v
							side := k%2 == 0
							var price uint64
							if side {
								price = buyPrices[k%5]
							} else {
								price = sellPrices[k%5]
							}
							_, tx, fees, err := issuer.c.GenerateTransactionManual(parser, nil, &actions.AddOrder{
								Pair:     pair,
								Quantity: v,
								Price:    price,
								Side:     side,
								 // ensure txs are unique
							}, factory, 0, tm)
							if err != nil {
								hutils.Outf("{{orange}}failed to generate:{{/}} %v\n", err)
								continue
							}
							transferFee = fees
							if err := issuer.d.RegisterTx(tx); err != nil {
								hutils.Outf("{{orange}}failed to register:{{/}} %v\n", err)
								continue
							}
							issuer.l.Lock()
							issuer.outstandingTxs++
							issuer.l.Unlock()
							inflight.Add(1)
							sent.Add(1)
						}
						// Determine how long to sleep
						dur := time.Since(start)
						sleep := amath.Max(1000-dur.Milliseconds(), 0)
						t.Reset(time.Duration(sleep) * time.Millisecond)
					case <-gctx.Done():
						return gctx.Err()
					case <-cctx.Done():
						return nil
					case <-signals:
						exiting.Do(func() {
							hutils.Outf("{{yellow}}exiting broadcast loop{{/}}\n")
							cancel()
						})
						return nil
					}
				}
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}

		// Wait for all issuers to finish
		hutils.Outf("{{yellow}}waiting for issuers to return{{/}}\n")
		dctx, cancel := context.WithCancel(ctx)
		go func() {
			// Send a dummy transaction if shutdown is taking too long (listeners are
			// expired on accept if dropped)
			t := time.NewTicker(15 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					hutils.Outf("{{yellow}}remaining:{{/}} %d\n", inflight.Load())
					_ = submitDummy(dctx, cli, tcli, key.PublicKey(), factory)
				case <-dctx.Done():
					return
				}
			}
		}()
		wg.Wait()
		cancel()

		// Return funds
		hutils.Outf("{{yellow}}returning funds to %s{{/}}\n", crypto.Address("clob", key.PublicKey()))
		var (
			returnedAvaxBalance uint64
			returnedUsdcBalance uint64
			returnsSent     int
		)
		for i := 0; i < numAccounts; i++ {
			avaxBalance := funds[accounts[i].PublicKey()][avaxID]
			if transferFee > avaxBalance {
				continue
			}
			returnsSent++
			// Send funds
			returnAmt := avaxBalance - transferFee
			_, tx, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    key.PublicKey(),
				TokenID: avaxID,
				Amount: returnAmt,
			}, auth.NewEIP712Factory(accounts[i]))
			if err != nil {
				return err
			}
			if err := dcli.RegisterTx(tx); err != nil {
				return err
			}
			returnedAvaxBalance += returnAmt

			usdcBalance := funds[accounts[i].PublicKey()][usdcID]
			if transferFee > avaxBalance {
				continue
			}
			returnsSent++
			returnAmt = usdcBalance - transferFee
			_, tx, _, err = cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    key.PublicKey(),
				TokenID: usdcID,
				Amount: returnAmt,
			}, auth.NewEIP712Factory(accounts[i]))
			if err != nil {
				return err
			}
			if err := dcli.RegisterTx(tx); err != nil {
				return err
			}
			returnedUsdcBalance += returnAmt

			// Ensure Snowman++ is activated
			if i < 10 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		for i := 0; i < returnsSent; i++ {
			_, dErr, result, err := dcli.ListenTx(ctx)
			if err != nil {
				return err
			}
			if dErr != nil {
				return dErr
			}
			if !result.Success {
				// Should never happen
				return errors.New("failed to return funds")
			}
		}
		hutils.Outf(
			"{{yellow}}returned funds:{{/}} %s %s\n",
			returnedAvaxBalance,
			avaxID,
		)
		hutils.Outf(
			"{{yellow}}returned funds:{{/}} %s %s\n",
			returnedUsdcBalance,
			usdcID,
		)
		return nil
	},
}
