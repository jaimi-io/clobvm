package orderbook

type OrderbookManager struct{
	orderbooks map[Pair]*Orderbook
}

func NewOrderbookManager() *OrderbookManager {
	return &OrderbookManager{
		orderbooks: make(map[Pair]*Orderbook),
	}
}

func (obm *OrderbookManager) GetOrderbook(pair Pair) *Orderbook {
	if ob, ok := obm.orderbooks[pair]; ok {
		return ob
	}
	ob := NewOrderbook()
	obm.orderbooks[pair] = ob
	return ob
}
