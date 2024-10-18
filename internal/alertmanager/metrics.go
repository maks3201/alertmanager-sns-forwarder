package alertmanager

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	AlertsReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sns_alerts_received_total",
			Help: "Total number of alerts received",
		},
	)

	AlertsFiltered = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sns_alerts_filtered_total",
			Help: "Total number of alerts filtered and not sent",
		},
	)

	AlertsSent = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sns_alerts_sent_total",
			Help: "Total number of alerts sent to AWS SNS",
		},
	)

	AlertsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sns_alerts_failed_total",
			Help: "Total number of alerts failed to send to AWS SNS",
		},
	)

	BatchesSent = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sns_batches_sent_total",
			Help: "Total number of alert batches sent to AWS SNS",
		},
	)

	SNSSendDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sns_send_duration_seconds",
			Help:    "Duration of sending alerts to AWS SNS",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	prometheus.MustRegister(AlertsReceived)
	prometheus.MustRegister(AlertsFiltered)
	prometheus.MustRegister(AlertsSent)
	prometheus.MustRegister(AlertsFailed)
	prometheus.MustRegister(BatchesSent)
	prometheus.MustRegister(SNSSendDuration)
}
