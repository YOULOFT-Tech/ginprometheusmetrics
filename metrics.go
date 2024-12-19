package ginprometheusmetrics

import "github.com/prometheus/client_golang/prometheus"

var (

	//example 0.1s~0.6s~1.1s~1.6s...4.6s~+Inf
	Interval500Mill  = prometheus.LinearBuckets(0.1, 0.5, 10)
	Interval1000Mill = prometheus.LinearBuckets(0.1, 1, 10)
	Interval2000Mill = prometheus.LinearBuckets(0.1, 2, 10)
)

func newMetric(ns string, dm DefineMetric) prometheus.Collector {

	var metric prometheus.Collector

	switch dm.MetricType {
	case "counter":
		metric = prometheus.NewCounterVec(

			prometheus.CounterOpts{
				Namespace: ns,
				Name:      dm.Name,
				Help:      dm.Help,
			},
			dm.Args,
		)

	case "gauge":
		metric = prometheus.NewGaugeVec(

			prometheus.GaugeOpts{
				Namespace: ns,
				Name:      dm.Name,
				Help:      dm.Help,
			},
			dm.Args,
		)

	case "histogram":
		metric = prometheus.NewHistogramVec(

			prometheus.HistogramOpts{
				Namespace: ns,
				Name:      dm.Name,
				Help:      dm.Help,
				Buckets:   dm.Buckets,
			},
			dm.Args,
		)

	case "summary":
		metric = prometheus.NewSummaryVec(

			prometheus.SummaryOpts{
				Namespace: ns,
				Name:      dm.Name,
				Help:      dm.Help,
			},
			dm.Args,
		)

	}

	return metric
}
