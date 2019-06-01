package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

const defaultNamespace string = "kafka_snitch"

type PrometheusExporter struct {
	namespace      string
	duration       prometheus.Gauge
	observations   prometheus.Counter
	brokerCount    prometheus.Gauge
	partitionCount prometheus.Gauge
	topicCount     prometheus.Gauge
	groupCount     prometheus.Gauge

	topicLag        *prometheus.GaugeVec
	partitionLag    *prometheus.GaugeVec
	partitionOffset *prometheus.GaugeVec

	mutex           sync.Mutex
	hasObservations bool
}

func NewPrometheusExporter(config *PrometheusConfig) (*PrometheusExporter, error) {
	namespace := defaultNamespace
	if config.Namespace != "" {
		namespace = config.Namespace
	}

	exporter := &PrometheusExporter{
		namespace: namespace,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "observation",
			Name:      "duration_seconds",
			Help:      "Duration of the last observation in seconds.",
		}),
		observations: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "observations_total",
			Help:      "Total number of times an observation was made.",
		}),
		brokerCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "observation",
			Name:      "broker_count",
			Help:      "Current number of observed brokers.",
		}),
		partitionCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "observation",
			Name:      "partition_count",
			Help:      "Current number of observed partitions.",
		}),
		topicCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "observation",
			Name:      "topic_count",
			Help:      "Current number of observed topics.",
		}),
		groupCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "observation",
			Name:      "consumer_group_count",
			Help:      "Current number of observed consumer groups.",
		}),
		topicLag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumer_topic_lag",
			Help:      "Current sum of consumer lag across a topic.",
		}, []string{"consumer_group", "topic"}),
		partitionOffset: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumer_partition_offset",
			Help:      "Current partition offset of a consumer.",
		}, []string{"consumer_group", "topic", "partition"}),
		partitionLag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumer_partition_lag",
			Help:      "Current delta between a partition's log end offset and consumer offset.",
		}, []string{"consumer_group", "topic", "partition"}),
	}

	prometheus.Register(exporter)
	if config.WebAddr != "" {
		go exporter.serve(config.WebAddr, config.WebPath)
	}

	return exporter, nil
}

func (pe *PrometheusExporter) serve(addr, path string) {
	http.Handle(path, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Kafka Snitch Exporter</title></head>
			<body>
			<h1>Kafka Snitch Exporter</h1>
			<p><a href="` + path + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Info("Starting Prometheus handler ", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (pe *PrometheusExporter) Describe(ch chan<- *prometheus.Desc) {
	log.Debugln("Describing metrics")

	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	ch <- pe.duration.Desc()
	ch <- pe.observations.Desc()
	if pe.hasObservations {
		ch <- pe.brokerCount.Desc()
		ch <- pe.topicCount.Desc()
		ch <- pe.partitionCount.Desc()
		ch <- pe.groupCount.Desc()
		pe.topicLag.Describe(ch)
		pe.partitionOffset.Describe(ch)
		pe.partitionLag.Describe(ch)
	}
}

func (pe *PrometheusExporter) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Prometheus is collecting metrics")

	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	ch <- pe.duration
	ch <- pe.observations
	if pe.hasObservations {
		ch <- pe.brokerCount
		ch <- pe.topicCount
		ch <- pe.partitionCount
		ch <- pe.groupCount
		pe.topicLag.Collect(ch)
		pe.partitionOffset.Collect(ch)
		pe.partitionLag.Collect(ch)
	}
}

func (pe *PrometheusExporter) WriteConsumerTopicLag(group, topic string, lag int64, tags Tags) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	pe.topicLag.With(prometheus.Labels{
		"consumer_group": group,
		"topic":          topic,
	}).Set(float64(lag))
}

func (pe *PrometheusExporter) WriteConsumerPartitionLag(group, topic string, partition int, logEndOffset, consumerOffset, lag int64, tags Tags) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	pe.partitionOffset.With(prometheus.Labels{
		"consumer_group": group,
		"topic":          topic,
		"partition":      strconv.Itoa(partition),
	}).Set(float64(consumerOffset))
	pe.partitionLag.With(prometheus.Labels{
		"consumer_group": group,
		"topic":          topic,
		"partition":      strconv.Itoa(partition),
	}).Set(float64(lag))
}

func (pe *PrometheusExporter) WriteObservationSummary(duration time.Duration, observationCount int64, brokerCount, topicCount, groupCount, partitionCount int, tags Tags) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	pe.observations.Inc()
	pe.duration.Set(duration.Seconds())
	pe.brokerCount.Set(float64(brokerCount))
	pe.topicCount.Set(float64(topicCount))
	pe.partitionCount.Set(float64(partitionCount))
	pe.groupCount.Set(float64(groupCount))

	pe.hasObservations = true
}

func (pe *PrometheusExporter) Flush() {
}
