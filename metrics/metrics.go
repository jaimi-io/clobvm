package metrics

import (
	"time"

	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/utils/metric"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/jaimi-io/clobvm/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	transfer      prometheus.Counter
	addOrder      prometheus.Counter
	cancelOrder   prometheus.Counter
	limitOrder    prometheus.Counter
	marketOrder   prometheus.Counter


	orderCancelNum   prometheus.Counter
	orderNum         prometheus.Gauge
	orderAmount      prometheus.Gauge
  orderFillsNum    prometheus.Counter
	orderFillsAmount prometheus.Counter
	orderProcessing  metric.Averager
}

func NewMetrics(gatherer ametrics.MultiGatherer) (*Metrics, error) {
	r := prometheus.NewRegistry()
	orderProcessing, err := metric.NewAverager(
		"orders",
		"order_processing",
		"order processing time",
		r,
	)
	if err != nil {
		return  nil, err
	}
	m := &Metrics{
		transfer: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "transfer",
			Help:      "number of transfer actions",
		}),
		addOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "add_order",
			Help:      "number of add order actions",
		}),
		cancelOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "cancel_order",
			Help:      "number of cancel order actions",
		}),
		limitOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "orders",
			Name:      "limit_order",
			Help:      "number of successful limit orders entered",
		}),
		marketOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "orders",
			Name:      "market_order",
			Help:      "number of successful market orders entered",
		}),
		orderCancelNum: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "orders",
			Name:      "order_cancel_num",
			Help:      "number of order cancels",
		}),
		orderNum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "orders",
			Name:      "order_num",
			Help:      "number of open orders",
		}),
		orderAmount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "orders",
			Name:      "order_amount",
			Help:      "sum of open orders",
		}),
		orderFillsNum: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "orders",
			Name:      "order_fills_num",
			Help:      "number of order fills",
		}),
		orderFillsAmount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "orders",
			Name:      "order_fills_amount",
			Help:      "sum of order fills",
		}),
		orderProcessing: orderProcessing,
	}
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.transfer),
		r.Register(m.addOrder),
		r.Register(m.cancelOrder),
		r.Register(m.limitOrder),
		r.Register(m.marketOrder),
		r.Register(m.orderCancelNum),
		r.Register(m.orderNum),
		r.Register(m.orderAmount),
		r.Register(m.orderFillsNum),
		r.Register(m.orderFillsAmount),
		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}

func (m *Metrics) Transfer() {
	m.transfer.Inc()
}

func (m *Metrics) AddOrder() {
	m.addOrder.Inc()
}

func (m *Metrics) CancelOrder() {
	m.cancelOrder.Inc()
}

func (m *Metrics) LimitOrder() {
	m.limitOrder.Inc()
}

func (m *Metrics) MarketOrder() {
	m.marketOrder.Inc()
}

func (m *Metrics) OrderCancelNum() {
	m.orderCancelNum.Inc()
}

func (m *Metrics) OrderFillsNum() {
	m.orderFillsNum.Inc()
}

func (m *Metrics) OrderFillsAmount(amount uint64) {
	m.orderFillsAmount.Add(float64(amount * utils.MinQuantity()))
}

func (m *Metrics) ObserverOrderProcessing(ts time.Duration) {
	m.orderProcessing.Observe(float64(ts))
}

func (m *Metrics) OrderNumInc() {
	m.orderNum.Inc()
}

func (m *Metrics) OrderNumDec() {
	m.orderNum.Dec()
}

func (m *Metrics) OrderAmountAdd(amount uint64) {
	m.orderAmount.Add(float64(amount * utils.MinQuantity()))
}

func (m *Metrics) OrderAmountSub(amount uint64) {
	m.orderAmount.Sub(float64(amount * utils.MinQuantity()))
}