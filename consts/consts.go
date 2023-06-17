package consts

import "time"

const (
	EvictionBlockWindow = uint64(1000)
	PendingBlockWindow  = uint64(7)
	ExecHistoryWindow   = 100 // s

	BalanceDecimals  = 9
	QuantityDecimals = 5
	PriceDecimals    = 4

	Day                     = time.Hour * 24
	NumExecutionHistoryDays = 30

	JSONRPCEndpoint = "/clobapi"
	Name            = "clobvm"
)