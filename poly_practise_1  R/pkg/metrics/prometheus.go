package metrics

import (
	"context"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	KafkaChannelUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_channel_usage",
			Help: "Current usage of Kafka channels",
		},
		[]string{"channel", "len", "cap"},
	)

	// Добавляем счетчики для producer и consumer
	ProducerMessagesSent = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "producer_messages_sent_total",
			Help: "Total number of messages sent by producer",
		},
	)

	ConsumerMessagesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "consumer_messages_received_total",
			Help: "Total number of messages received by consumer",
		},
	)

	// Gauge для lag
	MessageLag = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "message_lag",
			Help: "Difference between sent and received messages",
		},
	)

	ApplicationUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "application_uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)

	AppGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_goroutines",
			Help: "Number of application goroutines",
		},
	)

	AppThreads = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_threads",
			Help: "Number of application OS threads",
		},
	)

	MemoryAlloc = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_memory_alloc_bytes",
			Help: "Current memory allocation in bytes",
		},
	)

	MemoryHeapAlloc = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_memory_heap_alloc_bytes",
			Help: "Current heap memory allocation in bytes",
		},
	)

	MemoryHeapSys = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_memory_heap_sys_bytes",
			Help: "Total heap memory in bytes",
		},
	)

	GCCycles = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_gc_cycles_total",
			Help: "Total number of GC cycles",
		},
	)

	GCDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_gc_duration_seconds",
			Help:    "GC duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func UpdateKafkaChannelUsage(channelName string, length, capacity int) {
	KafkaChannelUsage.WithLabelValues(
		channelName,
		strconv.Itoa(length),
		strconv.Itoa(capacity),
	).Set(float64(length))
}

func UpdateRuntimeMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	MemoryAlloc.Set(float64(m.Alloc))
	MemoryHeapAlloc.Set(float64(m.HeapAlloc))
	MemoryHeapSys.Set(float64(m.HeapSys))

	AppGoroutines.Set(float64(runtime.NumGoroutine()))
	AppThreads.Set(float64(runtime.GOMAXPROCS(0)))

	GCCycles.Add(float64(m.NumGC))
	if m.NumGC > 0 {
		GCDuration.Observe(float64(m.PauseNs[(m.NumGC-1)%256]) / 1e9)
	}
	ApplicationUptime.Set(float64(time.Now().Unix()))
}

func SetupMetricsServer(addr string, logger *zap.Logger) {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				UpdateRuntimeMetrics()
			}
		}
	}()

	go func() {
		logger.Info("Starting metrics server", zap.String("address", addr))
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("Failed to start metrics server",
				zap.String("address", addr),
				zap.Error(err))
		}
	}()
}

// Serve starts the metrics server with context support
func Serve(ctx context.Context, addr string) error {
	http.Handle("/metrics", promhttp.Handler())

	// Start metrics update loop
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				UpdateRuntimeMetrics()
			}
		}
	}()

	// Start HTTP server
	server := &http.Server{Addr: addr}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("Failed to start metrics server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}
