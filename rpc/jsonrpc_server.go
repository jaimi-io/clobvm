package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/hypersdk/crypto"
)

type JSONRPCServer struct {
	c Controller
}

const JSONRPCEndpoint = "/clobapi"

func NewRPCServer(c Controller) *JSONRPCServer {
	return &JSONRPCServer{c}
}

type BalanceArgs struct {
	Address string `json:"address"`
	TokenID ids.ID `json:"tokenID"`
}

type BalanceReply struct {
	Balance uint64 `json:"balance"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Balance")
	defer span.End()
	
	address, err := crypto.ParseAddress("clob", args.Address)
	if err != nil {
		return err
	}
	bal, err := j.c.GetBalance(ctx, address, args.TokenID)
	if err != nil {
		return err
	}
	reply.Balance = bal
	return nil
}

type AllOrdersArgs struct {}
type AllOrdersReply struct {
	BuySide string `json:"buySide"`
	SellSide string `json:"sellSide"`
}
func (j *JSONRPCServer) AllOrders(req *http.Request, args *AllOrdersArgs, reply *AllOrdersReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.AllOrders")
	defer span.End()

	var err error
	reply.BuySide, reply.SellSide, err = j.c.GetOrderbook(ctx)
	return err
}