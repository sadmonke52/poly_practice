package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var reg = prometheus.NewRegistry()

var (
	MessagesConsumed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Total number of messages consumed from Kafka",
		},
	)

	MessagesFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_failed_total",
			Help: "Total number of failed messages from Kafka",
		},
		[]string{"reason"},
	)

	InFlightMessages = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "in_flight_messages",
			Help: "Number of messages currently being processed",
		},
	)

	KafkaConsumerLag = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag",
			Help: "Kafka consumer lag (latest offset - current offset)",
		},
	)

	QueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "internal_queue_size",
			Help: "Length of internal processing queue/channel",
		},
		[]string{"stage"},
	)
)

func Init() {

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	reg.MustRegister(MessagesConsumed)
	reg.MustRegister(MessagesFailed)
	reg.MustRegister(InFlightMessages)
	reg.MustRegister(KafkaConsumerLag)
	reg.MustRegister(QueueSize)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
