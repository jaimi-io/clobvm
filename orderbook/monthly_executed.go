package orderbook

import (
	"container/ring"
	"time"

	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/hypersdk/crypto"
)

type Execution struct {
	Timestamp int64
	Quantity  uint64
}

type MonthlyExecuted struct {
	executions *ring.Ring
	total  uint64
}

func NewMonthlyExecuted(timestamp int64) *MonthlyExecuted {
	// +1 to store current day (doesn't count towards total)
	executionsRing := ring.New(consts.NumExecutionHistoryDays + 1)
	executionsRing.Value = &Execution{
		Timestamp: time.Unix(timestamp, 0).Truncate(consts.Day).UnixMilli(),
		Quantity:  0,
	}
	return &MonthlyExecuted{
		executions: executionsRing,
	}
}

func (me *MonthlyExecuted) advanceExecs(blockTs int64) {
	bufferWindow := time.Second * consts.ExecHistoryWindow
	lastSyncTs := time.Unix(blockTs, 0).Add(-bufferWindow).Truncate(consts.Day).UnixMilli()
	currentNode := me.executions
	currentExec := currentNode.Value.(*Execution)
	currentTs := currentExec.Timestamp

	// TODO: optimise if currentTs > lastSyncTs - 31 days
	
	for currentTs < lastSyncTs {
		// Add current day total to monthly total
		me.total += currentExec.Quantity

		// Advance to next day
		currentNode = currentNode.Next()
		currentTs = time.Unix(currentTs, 0).Add(consts.Day).UnixMilli()
		if currentNode.Value == nil {
			currentNode.Value = &Execution{
				Timestamp: currentTs,
				Quantity:  0,
			}
		} else {
			// Reset next day - 31 days ago from monthly total, set to be current day
			currentExec = currentNode.Value.(*Execution)
			me.total -= currentExec.Quantity
			currentExec.Quantity = 0
			currentExec.Timestamp = currentTs
		}
	}

	me.executions = currentNode
}

func (me *MonthlyExecuted) AddExec(timestamp int64, quantity uint64) {
	me.advanceExecs(timestamp)
	exec := me.executions.Value.(*Execution)
	exec.Quantity += quantity
}

func (me *MonthlyExecuted) getMonthlyExecuted(blockTs int64) uint64 {
	me.advanceExecs(blockTs)
	return me.total
}

func (ob *Orderbook) addExec(user crypto.PublicKey, timestamp int64, quantity uint64) {
	if _, ok := ob.executionHistory[user]; !ok {
		ob.executionHistory[user] = NewMonthlyExecuted(timestamp)
	}
	ob.executionHistory[user].AddExec(timestamp, quantity)
}

func (ob *Orderbook) GetFee(user crypto.PublicKey, timestamp int64, quantity uint64) uint64 {
	if _, ok := ob.executionHistory[user]; !ok {
		return CalculateTakerFee(0, quantity)
	}
	monthlyExecuted := ob.executionHistory[user].getMonthlyExecuted(timestamp)
	return CalculateTakerFee(monthlyExecuted, quantity)
}

func (ob *Orderbook) GetFeeRate(user crypto.PublicKey, timestamp int64) float64 {
	if _, ok := ob.executionHistory[user]; !ok {
		return GetMakerRate(0)
	}
	monthlyExecuted := ob.executionHistory[user].getMonthlyExecuted(timestamp)
	return GetMakerRate(monthlyExecuted)
}

func (ob *Orderbook) RefundFee(user crypto.PublicKey, timestamp int64, quantity uint64) uint64 {
	if _, ok := ob.executionHistory[user]; !ok {
		return RefundTakerFee(0, quantity)
	}
	monthlyExecuted := ob.executionHistory[user].getMonthlyExecuted(timestamp)
	return RefundTakerFee(monthlyExecuted, quantity)
}