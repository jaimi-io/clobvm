package cmd

import (
	"context"
	"errors"
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
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/pubsub"
	"github.com/jaimi-io/hypersdk/rpc"
	hutils "github.com/jaimi-io/hypersdk/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type marketMaker struct {
	buyPrices       []uint64
	sellPrices      []uint64
	midPrice       uint64
	orders         []ids.ID
	ordersLock     sync.Mutex
	orderCounter   uint64
}

var (
	pair orderbook.Pair
	inflight atomic.Int64
	marketOrder atomic.Int64
	expired atomic.Int64
	sent atomic.Int64
	exiting sync.Once
	numTransactions int
	marketMakerQuantity uint64
	curSpreads map[float64]struct{}
)

func (m *marketMaker) UpdateParams(newMidPrice uint64) bool {
	if newMidPrice == m.midPrice {
		return false
	}
	m.midPrice = newMidPrice
	m.buyPrices, m.sellPrices = calculateSpread(newMidPrice)
	return true
}

func calculateSpread(midPrice uint64) ([]uint64, []uint64) {
	desiredStdDev := 0.25
	desiredMean := 0.5
	buyPrices := make([]uint64, 0)
	sellPrices := make([]uint64, 0)

	for i := 0; i < numTransactions; i++ {
		random := rand.New(rand.NewSource(time.Now().UnixNano()))
		spread := random.NormFloat64()*desiredStdDev + desiredMean
		for _, ok := curSpreads[spread]; spread < 0.1 && !ok ; _, ok = curSpreads[spread]{
			spread = random.NormFloat64()*desiredStdDev + desiredMean
		}
		dist := uint64(math.Round(float64(midPrice)*spread/100))
		buyPrices = append(buyPrices, midPrice-dist)
		sellPrices = append(sellPrices, midPrice+dist)
		curSpreads[spread] = struct{}{}
	}
	return buyPrices, sellPrices
}

func calculateMarketOrder(midPrice uint64) (uint64, bool) {
	desiredStdDev := float64(100)
	desiredMean := float64(10)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	quantity := random.NormFloat64()*desiredStdDev + desiredMean
	for quantity < 1 {
		quantity = random.NormFloat64()*desiredStdDev + desiredMean
	}
	qty := uint64(quantity * float64(utils.MinPrice())) * utils.MinQuantity()
	side := random.Intn(2) == 0
	return qty, side
}

func getMarketOrderSleep() (int64) {
	desiredStdDev := float64(5)
	desiredMean := float64(5)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	sleepSec := random.NormFloat64()*desiredStdDev + desiredMean
	for sleepSec < 1 {
		sleepSec = random.NormFloat64()*desiredStdDev + desiredMean
	}
	
	sleepMilli := int64(math.Round(sleepSec * 1000))
	return sleepMilli
}

func simulateMarketOrderer(issuer *txIssuer, parser chain.Parser, factory *auth.EIP712Factory, tm *timeModifier, localMidPrice uint64) {
	quantity, side := calculateMarketOrder(localMidPrice)
	_, tx, _, err := issuer.c.GenerateTransactionManual(parser, nil, &actions.AddOrder{
		Pair:     pair,
		Quantity: quantity,
		Side:     side,
		// ensure txs are unique
	}, factory, 0, tm)
	if err != nil {
		hutils.Outf("{{orange}}failed to generate:{{/}} %v\n", err)
		return
	}
	if err := issuer.d.RegisterTx(tx); err != nil {
		hutils.Outf("{{orange}}failed to register:{{/}} %v\n", err)
		return
	}
	issuer.l.Lock()
	issuer.outstandingTxs++
	issuer.l.Unlock()
	inflight.Add(1)
	sent.Add(1)
	marketOrder.Add(1)
}

func cancelAllOrders(issuer *txIssuer, parser chain.Parser, factory *auth.EIP712Factory, tm *timeModifier) {
	_, tx, _, err := issuer.c.GenerateTransactionManual(parser, nil, &actions.CancelOrder{
		Pair:     pair,
		OrderID: ids.GenerateTestID(),
	}, factory, 0, tm)
	if err != nil {
		hutils.Outf("{{orange}}failed to generate:{{/}} %v\n", err)
		return
	}
	if err := issuer.d.RegisterTx(tx); err != nil {
		hutils.Outf("{{orange}}failed to register:{{/}} %v\n", err)
		return
	}
	issuer.l.Lock()
	issuer.outstandingTxs++
	issuer.l.Unlock()
	inflight.Add(1)
	sent.Add(1)
}

func simulateMarketMaker(issuer *txIssuer, parser chain.Parser, factory *auth.EIP712Factory, tm *timeModifier, localMidPrice uint64, i int, mm *marketMaker) {
	toUpdate := mm.UpdateParams(localMidPrice)
	if toUpdate {
		cancelAllOrders(issuer, parser, factory, tm)
	}
	for k := 0; k < 2 * numTransactions; k++ {
		side := (k+i)%2 == 0
		var price uint64
		if side {
			price = mm.buyPrices[k/2]
		} else {
			price = mm.sellPrices[k/2]
		}
		mm.orderCounter += 1
		_, tx, _, err := issuer.c.GenerateTransactionManual(parser, nil, &actions.AddOrder{
			Pair:     pair,
			Quantity: marketMakerQuantity,
			Price:    price,
			Side:     side,
			// ensure txs are unique
		}, factory, 0, tm)
		if err != nil {
			hutils.Outf("{{orange}}failed to generate:{{/}} %v\n", err)
			continue
		}
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
}

var simulateOrderCmd = &cobra.Command{
	Use: "simulate",
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
		pair = orderbook.Pair{
			BaseTokenID:  avaxID,
			QuoteTokenID: usdcID,
		}
		marketMakerQuantity = 20 * utils.MinBalance()
		midPrice := 100 * utils.MinPrice()

		// Distribute funds
		numMarketMakers, err := promptInt("number of market makers")
		if err != nil {
			return err
		}
		numTransactions, err = promptInt("number of symmetrical orders")
		if err != nil {
			return err
		}
		numAccounts := numMarketMakers + numTransactions
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

		marketMakers := make(map[crypto.PublicKey]*marketMaker)
		curSpreads = make(map[float64]struct{})
		for i := 0; i < numMarketMakers; i++ {
			buyPrices, sellPrices := calculateSpread(midPrice)
			marketMakers[accounts[i].PublicKey()] = &marketMaker{
				buyPrices:  buyPrices,
				sellPrices: sellPrices,
				midPrice:  midPrice,
				orders: make([]ids.ID, 0),
			}
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
							if balUp.BaseBalance > 0 {
								funds[balUp.BaseTokenUser][avaxID] = balUp.BaseBalance
							}
							if balUp.QuoteBalance > 0 {
								funds[balUp.QuoteTokenUser][usdcID] = balUp.QuoteBalance
							}
							fundsL.Unlock()
							if result.Units > 0 {
								_, ok := marketMakers[balUp.BaseTokenUser]
								if !ok {
									mid, _ := tcli.MidPrice(ctx, pair)
									if mid != 0 {
										midPrice = uint64(math.Round(mid * float64(utils.MinPrice())))
									}
								}
							}
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
							"{{yellow}}txs seen:{{/}} %d {{yellow}}success rate:{{/}} %.2f%% {{yellow}}inflight:{{/}} %d {{yellow}}issued/s:{{/}} %d {{yellow}}marketOrders:{{/}} %d\n", //nolint:lll
							totalTxs,
							float64(confirmedTxs)/float64(totalTxs)*100,
							inflight.Load(),
							current-psent,
							marketOrder.Load(),
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
						tm := &timeModifier{nextTime + parser.Rules(nextTime).GetValidityWindow() - 3}
						//midL.Lock()
						localMidPrice := midPrice
						//midL.Unlock()
						sleepMs := int64(1000)

						if i >= numMarketMakers {
							simulateMarketOrderer(issuer, parser, factory, tm, localMidPrice)
							sleepMs = getMarketOrderSleep()
						} else {
							mm := marketMakers[accounts[i].PublicKey()]
							simulateMarketMaker(issuer, parser, factory, tm, localMidPrice, i, mm)
						}

						// Determine how long to sleep
						dur := time.Since(start)
						sleep := amath.Max(sleepMs-dur.Milliseconds(), 0)
						if sleepMs-dur.Milliseconds() < 0 {
							hutils.Outf("{{red}}tx took too long:{{/}} %dms\n", dur.Milliseconds())
						}
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
