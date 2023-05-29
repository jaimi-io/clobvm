package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
)

type JSONRPCServer struct {

}

const JSONRPCEndpoint = "/clobapi"

func New() *JSONRPCServer {
	return &JSONRPCServer{}
}

type BalanceArgs struct {
	Address string `json:"address"`
	TokenID ids.ID `json:"tokenID"`
}

type BalanceReply struct {
	Balance uint64 `json:"balance"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	//address, _ := crypto.ParseAddress("clob", args.Address)
	reply.Balance = 0
	return nil
}