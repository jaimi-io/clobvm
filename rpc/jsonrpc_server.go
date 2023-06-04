package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/orderbook"
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

type AllOrdersArgs struct {
	Pair orderbook.Pair `json:"pair"`
}
type AllOrdersReply struct {
	BuySide  string `json:"buySide"`
	SellSide string `json:"sellSide"`
}
func (j *JSONRPCServer) AllOrders(req *http.Request, args *AllOrdersArgs, reply *AllOrdersReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.AllOrders")
	defer span.End()

	var err error
	reply.BuySide, reply.SellSide, err = j.c.GetOrderbook(ctx, args.Pair)
	return err
}

type PendingFundsArgs struct {
	User        crypto.PublicKey `json:"user"`
	TokenID     ids.ID           `json:"tokenID"`
	BlockHeight uint64           `json:"blockHeight"`
}
type PendingFundsReply struct {
	Balance uint64     `json:"balance"`
	BlockHeight uint64 `json:"blockHeight"`
}
func (j *JSONRPCServer) PendingFunds(req *http.Request, args *PendingFundsArgs, reply *PendingFundsReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.PendingFunds")
	defer span.End()

	var err error
	reply.Balance, reply.BlockHeight = j.c.GetPendingFunds(ctx, args.User, args.TokenID, args.BlockHeight)
	return err
}

type VolumesArgs struct {
	Pair orderbook.Pair `json:"pair"`
}
type VolumesReply struct {
	Volumes  string `json:"volumes"`
}
func (j *JSONRPCServer) Volumes(req *http.Request, args *VolumesArgs, reply *VolumesReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Volumes")
	defer span.End()

	var err error
	reply.Volumes, err = j.c.GetVolumes(ctx, args.Pair)
	return err
}