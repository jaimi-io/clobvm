package rpc

import (
	"context"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/clobvm/registry"

	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/requester"
)

type JSONRPCClient struct {
	requester *requester.EndpointRequester
	chainID ids.ID
	genesis *genesis.Genesis
}

func NewRPCClient(uri string, chainID ids.ID, g *genesis.Genesis) *JSONRPCClient {
	uri = strings.TrimSuffix(uri, "/")
	uri += JSONRPCEndpoint
	req := requester.New(uri, "clobvm")
	return &JSONRPCClient{req, chainID, g}
}

func (j *JSONRPCClient) Balance(ctx context.Context, address string, tokenID ids.ID) (uint64, error) {
	args := &BalanceArgs{
		Address: address,
		TokenID: tokenID,
	}
	var reply BalanceReply
	err := j.requester.SendRequest(ctx, "balance", args, &reply)
	return reply.Balance, err
}


type Parser struct {
	chainID ids.ID
	genesis *genesis.Genesis
}

func (p *Parser) ChainID() ids.ID {
	return p.chainID
}

func (p *Parser) Rules(t int64) chain.Rules {
	return &genesis.Rules{}
}

func (*Parser) Registry() (chain.ActionRegistry, chain.AuthRegistry) {
	return registry.ActionRegistry, registry.AuthRegistry
}

func (cli *JSONRPCClient) Parser(ctx context.Context) (chain.Parser, error) {
	return &Parser{cli.chainID, cli.genesis}, nil
}
