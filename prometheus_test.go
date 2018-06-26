package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_prometheus_exporter(t *testing.T) {
	exporter, _ := NewPrometheusExporter(&PrometheusConfig{})

	//prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	server := httptest.NewServer(prometheus.UninstrumentedHandler())
	defer server.Close()

	exporter.WriteConsumerTopicLag("Istanbul", "works", 1, Tags{})
	exporter.WriteConsumerPartitionLag("Constantinople", "works", 1, 2, 3, 4, Tags{})
	exporter.WriteObservationSummary(10*time.Second, 1, 2, 3, 4, 5, Tags{})

	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	scrape := string(body)
	if line := `kafka_snitch_observations_total 1`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_observation_duration_seconds 10`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_observation_broker_count 2`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_observation_topic_count 3`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_observation_consumer_group_count 4`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_observation_partition_count 5`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_consumer_topic_lag{consumer_group="Istanbul",topic="works"} 1`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_consumer_partition_lag{consumer_group="Constantinople",partition="1",topic="works"} 4`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
	if line := `kafka_snitch_consumer_partition_offset{consumer_group="Constantinople",partition="1",topic="works"} 3`; !strings.Contains(scrape, line) {
		t.Errorf("No metric matching: %s\n", line)
	}
}
