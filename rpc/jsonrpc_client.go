package rpc

import (
	"context"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/orderbook"
	"github.com/jaimi-io/clobvm/registry"

	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/requester"
)

type JSONRPCClient struct {
	requester *requester.EndpointRequester
	chainID ids.ID
	genesis *genesis.Genesis
}

func NewRPCClient(uri string, chainID ids.ID) *JSONRPCClient {
	uri = strings.TrimSuffix(uri, "/")
	uri += consts.JSONRPCEndpoint
	req := requester.New(uri, "clobvm")
	return &JSONRPCClient{req, chainID, nil}
}

func (cli *JSONRPCClient) Genesis(ctx context.Context) (*genesis.Genesis, error) {
	if cli.genesis != nil {
		return cli.genesis, nil
	}

	resp := new(GenesisReply)
	err := cli.requester.SendRequest(
		ctx,
		"genesis",
		nil,
		resp,
	)
	if err != nil {
		return nil, err
	}
	cli.genesis = resp.Genesis
	return resp.Genesis, nil
}

func (j *JSONRPCClient) Balance(ctx context.Context, address string, tokenID ids.ID) (float64, error) {
	args := &BalanceArgs{
		Address: address,
		TokenID: tokenID,
	}
	var reply BalanceReply
	err := j.requester.SendRequest(ctx, "balance", args, &reply)
	return reply.Balance, err
}

func (j *JSONRPCClient) MidPrice(ctx context.Context, pair orderbook.Pair) (float64, error) {
	args := &MidPriceArgs{
		Pair: pair,
	}
	var reply MidPriceReply
	err := j.requester.SendRequest(ctx, "midPrice", args, &reply)
	return reply.MidPrice, err
}

func (j *JSONRPCClient) AllOrders(ctx context.Context, pair orderbook.Pair, numPriceLevels int) (string, string, error) {
	args := &AllOrdersArgs{
		Pair: pair,
		NumPriceLevels: numPriceLevels,
	}
	var reply AllOrdersReply
	err := j.requester.SendRequest(ctx, "allOrders", args, &reply)
	return reply.BuySide, reply.SellSide, err
}

func (j *JSONRPCClient) PendingFunds(ctx context.Context, user crypto.PublicKey, tokenID ids.ID, blockHeight uint64) (float64, uint64, error) {
	args := &PendingFundsArgs{
		User:        user,
		TokenID:     tokenID,
		BlockHeight: blockHeight,
	}
	var reply PendingFundsReply
	err := j.requester.SendRequest(ctx, "pendingFunds", args, &reply)
	return reply.Balance, reply.BlockHeight, err
}

func (j *JSONRPCClient) Volumes(ctx context.Context, pair orderbook.Pair, numPriceLevels int) (string, error) {
	args := &VolumesArgs{
		Pair: pair,
		NumPriceLevels: numPriceLevels,
	}
	var reply VolumesReply
	err := j.requester.SendRequest(ctx, "volumes", args, &reply)
	return reply.Volumes, err
}

type Parser struct {
	chainID ids.ID
	genesis *genesis.Genesis
}

func (p *Parser) ChainID() ids.ID {
	return p.chainID
}

func (p *Parser) Rules(t int64) chain.Rules {
	return p.genesis.NewRules()
}

func (*Parser) Registry() (chain.ActionRegistry, chain.AuthRegistry) {
	return registry.ActionRegistry, registry.AuthRegistry
}

func (cli *JSONRPCClient) Parser(ctx context.Context) (chain.Parser, error) {
	g, err := cli.Genesis(ctx)
	if err != nil {
		return nil, err
	}
	return &Parser{cli.chainID, g}, nil
}
