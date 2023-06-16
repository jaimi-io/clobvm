package controller

import (
	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/jaimi-io/clobvm/consts"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	transfer      prometheus.Counter
	addOrder      prometheus.Counter
	cancelOrder   prometheus.Counter
}

func NewMetrics(gatherer ametrics.MultiGatherer) (*Metrics, error) {
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
	}
	r := prometheus.NewRegistry()
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.transfer),

		r.Register(m.addOrder),
		r.Register(m.cancelOrder),
		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}
