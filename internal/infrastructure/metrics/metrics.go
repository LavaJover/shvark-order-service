package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OrdersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Total number of orders",
		},
		[]string{"status", "reason"}, // status=succeed|failed
	)
)

func Init() {
	prometheus.MustRegister(OrdersTotal)
}

func IncOrder(status, reason string){
	OrdersTotal.WithLabelValues(status, reason).Inc()
}