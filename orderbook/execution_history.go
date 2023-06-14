package orderbook

import (
	"time"

	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Exec struct {
	Timestamp int64
	Quantity  uint64
}

type ExecHistory struct {
	executions      []*Exec
	monthlyExecuted uint64
	latestTimestamp int64
}

func NewExecHistory() *ExecHistory {
	return &ExecHistory{
		executions: make([]*Exec, 0),
	}
}

func (eh *ExecHistory) removeOldExecs(blockTs int64) {
	month := time.Hour * 24 * 30
	bufferWindow := time.Second * consts.ExecHistoryWindow
	prevMonthTs := time.Unix(blockTs, 0).Add(-month-bufferWindow).Unix()
	for len(eh.executions) > 0 {
		if eh.executions[0].Timestamp < prevMonthTs {
			eh.monthlyExecuted -= eh.executions[0].Quantity
			eh.executions = eh.executions[1:]
		} else {
			break
		}
	}
	if len(eh.executions) == 0 {
		eh.latestTimestamp = 0
	}
}

func (eh *ExecHistory) advanceExecs(blockTs int64) {
	bufferWindow := time.Second * consts.ExecHistoryWindow
	latestValidTs := time.Unix(blockTs, 0).Add(-bufferWindow).UnixMilli()
	prevStoredTs := eh.latestTimestamp
	for i := len(eh.executions)-1; i >= 0; i-- {
		if eh.executions[i].Timestamp == eh.latestTimestamp {
			break
		}
		if eh.executions[i].Timestamp <= latestValidTs {
			if prevStoredTs == eh.latestTimestamp {
				eh.latestTimestamp = eh.executions[i].Timestamp
			}
			eh.monthlyExecuted += eh.executions[i].Quantity
		}
	}
}

func (eh *ExecHistory) AddExec(timestamp int64, quantity uint64) {
	if n := len(eh.executions); n > 0 && eh.executions[n-1].Timestamp == timestamp {
		eh.executions[n-1].Quantity += quantity
	} else {
		eh.executions = append(eh.executions, &Exec{timestamp, quantity})
	}
}

func (eh *ExecHistory) getMonthlyExecuted(blockTs int64) uint64 {
	eh.removeOldExecs(blockTs)
	eh.advanceExecs(blockTs)
	return eh.monthlyExecuted
}

func (ob *Orderbook) addExec(user crypto.PublicKey, timestamp int64, quantity uint64) {
	if _, ok := ob.executionHistory[user]; !ok {
		ob.executionHistory[user] = NewExecHistory()
	}
	ob.executionHistory[user].AddExec(timestamp, quantity)
}

func (ob *Orderbook) GetFee(user crypto.PublicKey, timestamp int64, quantity uint64) uint64 {
	if _, ok := ob.executionHistory[user]; !ok {
		return CalculateFee(0, quantity)
	}
	monthlyExecuted := ob.executionHistory[user].getMonthlyExecuted(timestamp)
	return CalculateFee(monthlyExecuted, quantity)
}

func (ob *Orderbook) GetFeeRate(user crypto.PublicKey, timestamp int64) float64 {
	if _, ok := ob.executionHistory[user]; !ok {
		return GetFeeRate(0)
	}
	monthlyExecuted := ob.executionHistory[user].getMonthlyExecuted(timestamp)
	return GetFeeRate(monthlyExecuted)
}