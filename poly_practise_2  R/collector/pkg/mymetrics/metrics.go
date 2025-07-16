package mymetrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var reg = prometheus.NewRegistry()

var (
	MessagesConsumed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Total number of messages consumed from Kafka",
		},
		[]string{"topic"},
	)

	MessagesFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_failed_total",
			Help: "Total number of failed messages from Kafka",
		},
		[]string{"topic", "reason"},
	)

	InFlightMessages = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "in_flight_messages",
			Help: "Number of messages currently being processed",
		},
		[]string{"topic"},
	)

	KafkaConsumerLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag",
			Help: "Kafka consumer lag (latest offset - current offset)",
		},
		[]string{"topic"},
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
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})) //когда берём эти метрики происходит двойной вызов

	reg.MustRegister(MessagesConsumed)
	reg.MustRegister(MessagesFailed)
	reg.MustRegister(InFlightMessages)
	reg.MustRegister(KafkaConsumerLag)
	reg.MustRegister(QueueSize)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
