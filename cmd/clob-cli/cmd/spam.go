package cmd

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/genesis"
	trpc "github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/rpc"
	hutils "github.com/jaimi-io/hypersdk/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	feePerTx     = 1000
	defaultRange = 32
)

type txIssuer struct {
	c  *rpc.JSONRPCClient
	tc *trpc.JSONRPCClient
	d  *rpc.WebSocketClient

	l              sync.Mutex
	outstandingTxs int
}

var spamCmd = &cobra.Command{
	Use: "spam",
	RunE: func(*cobra.Command, []string) error {
		return errors.New("must specify a subcommand")
	},
}

func getRandomRecipient(self int, keys []crypto.PrivateKey) (crypto.PublicKey, error) {
	// if randomRecipient {
	priv, err := crypto.GeneratePrivateKey()
	if err != nil {
		return crypto.EmptyPublicKey, err
	}
	return priv.PublicKey(), nil
	// }

	// Select item from array
	// index := rand.Int() % len(keys)
	// if index == self {
	// 	index++
	// 	if index == len(keys) {
	// 		index = 0
	// 	}
	// }
	// return keys[index].PublicKey(), nil
}

func getRandomIssuer(issuers []*txIssuer) *txIssuer {
	index := rand.Int() % len(issuers)
	return issuers[index]
}

var runSpamCmd = &cobra.Command{
	Use: "run",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		chainID, key, _, _, _, err := defaultActor()
		if err != nil {
			return err
		}
		// NodeID-JXNpkB63kC6azj216DTV8U9CyeSs8pZoc URI: http://127.0.0.1:60622/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb
		// NodeID-36HK72jQ2TZf96KfFzBRS2QPJa15VFVGi URI: http://127.0.0.1:51007/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb
		// NodeID-PqpeTWZpxHa8AUu1YZLRXoM17dJe5s85v URI: http://127.0.0.1:23667/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb
		// NodeID-6925LBFyRcVHxKgvJaMorTVijeHp2Y5bN URI: http://127.0.0.1:27671/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb
		// NodeID-F1Pwj4rMjbRYYQKBjMPVUKT8hnRxzjkmU URI: http://127.0.0.1:43548/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb
		uris := []string{
			"http://127.0.0.1:60622/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb",
			"http://127.0.0.1:51007/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb",
			"http://127.0.0.1:23667/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb",
			"http://127.0.0.1:27671/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb",
			"http://127.0.0.1:43548/ext/bc/S9sfqYbsvpoRx2Sr4aAk3XkJESWrTVVau4kEALBj6CFrtmuMb",
		}

		// Select root key
		// keys, err := GetKeys()
		// if err != nil {
		// 	return err
		// }

		cli := rpc.NewJSONRPCClient(uris[0])
		tcli := trpc.NewRPCClient(uris[0], chainID, genesis.New())
		// hutils.Outf("{{cyan}}stored keys:{{/}} %d\n", len(keys))
		// balances := make([]uint64, len(keys))
		// for i := 0; i < len(keys); i++ {
		// 	address := crypto.Address("clob", keys[i].PublicKey())
		// 	balance, err := tcli.Balance(ctx, address, ids.Empty)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	balances[i] = balance
		// 	hutils.Outf(
		// 		"%d) {{cyan}}address:{{/}} %s {{cyan}}balance:{{/}} %s TKN\n",
		// 		i,
		// 		address,
		// 		valueString(ids.Empty, balance),
		// 	)
		// }
		// keyIndex, err := promptChoice("select root key", len(keys))
		// if err != nil {
		// 	return err
		// }
		balance := uint64(10000000)
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
		dcli, err := rpc.NewWebSocketClient(uris[0])
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
			funds[pk.PublicKey()] = distAmount

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
				return errors.New("tx failed")
			}
		}
		hutils.Outf("{{yellow}}distributed funds to %d accounts{{/}}\n", numAccounts)

		// Kickoff txs
		clients := make([]*txIssuer, len(uris))
		for i := 0; i < len(uris); i++ {
			cli := rpc.NewJSONRPCClient(uris[i])
			tcli := trpc.NewRPCClient(uris[i], chainID, genesis.New())
			dcli, err := rpc.NewWebSocketClient(uris[i])
			if err != nil {
				return err
			}
			clients[i] = &txIssuer{c: cli, tc: tcli, d: dcli}
		}
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
		unitPrice, _, err := clients[0].c.SuggestedRawFee(ctx)
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
				fundsL.Lock()
				balance := funds[accounts[i].PublicKey()]
				fundsL.Unlock()
				defer func() {
					fundsL.Lock()
					funds[accounts[i].PublicKey()] = balance
					fundsL.Unlock()
				}()
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
							}, factory, unitPrice)
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
						sleep := math.Max(1000-dur.Milliseconds(), 0)
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
