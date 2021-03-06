package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PagerDuty/godspeed"
)

const TagFmtDataDog = "datadog"
const TagFmtNone = "none"

var (
	statsdBadChars     = regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	statsdReplaceChars = strings.NewReplacer(
		" ", "-",
		"/", "-",
		"@", "-",
		"*", "-",
	)
	statsdRemoveChars = strings.NewReplacer(
		`'`, "",
		`"`, "",
		`\`, "",
		"..", ".",
	)
)

type StatsDWriter struct {
	gsw       *godspeed.Godspeed
	tagFormat string
}

func NewStatsDWriter(config *StatsDConfig) (*StatsDWriter, error) {
	var port int
	var host string
	var err error

	addr := strings.SplitN(config.Addr, ":", 2)
	if len(addr) == 1 {
		host = addr[0]
		port = 8125
	} else if len(addr) == 2 {
		host = addr[0]
		port, err = strconv.Atoi(addr[1])
		if err != nil {
			return nil, fmt.Errorf("Invalid host:port addr: %v", err)
		}
	}

	if host == "" {
		host = "localhost"
	}

	gs, err := godspeed.New(host, port, false)
	if err != nil {
		return nil, err
	}

	gs.Namespace = "kafka.snitch"

	return &StatsDWriter{
		gsw:       gs,
		tagFormat: config.TagFormat,
	}, nil
}

func (w *StatsDWriter) WriteConsumerTopicLag(group, topic string, lag int64, tags Tags) {
	if w.tagFormat == TagFmtDataDog {
		tags["consumer_group"] = group
		tags["topic"] = topic

		tagArray := []string{}
		for name, value := range tags {
			tagArray = append(tagArray, fmt.Sprintf("%s:%s", name, value))
		}

		w.gsw.Gauge("consumer.topic.lag", float64(lag), tagArray)
		return
	}

	w.gsw.Gauge(fmt.Sprintf("consumer.%s.topic.%s.lag", statsdSafeString(group), statsdSafeString(topic)), float64(lag), nil)
}

func (w *StatsDWriter) WriteConsumerPartitionLag(group, topic string, partition int, logEndOffset, consumerOffset, lag int64, tags Tags) {
	if w.tagFormat == TagFmtDataDog {
		tags["consumer_group"] = group
		tags["topic"] = topic
		tags["partition"] = strconv.Itoa(partition)

		tagArray := []string{}
		for name, value := range tags {
			tagArray = append(tagArray, fmt.Sprintf("%s:%s", name, value))
		}

		w.gsw.Gauge("consumer.partition.log_end_offset", float64(logEndOffset), tagArray)
		w.gsw.Gauge("consumer.partition.consumer_offset", float64(consumerOffset), tagArray)
		w.gsw.Gauge("consumer.partition.lag", float64(lag), tagArray)
		return
	}

	w.gsw.Gauge(fmt.Sprintf("consumer.%s.topic.%s.partition.%d.log_end_offset", statsdSafeString(group), statsdSafeString(topic), partition), float64(logEndOffset), nil)
	w.gsw.Gauge(fmt.Sprintf("consumer.%s.topic.%s.partition.%d.consumer_offset", statsdSafeString(group), statsdSafeString(topic), partition), float64(consumerOffset), nil)
	w.gsw.Gauge(fmt.Sprintf("consumer.%s.topic.%s.partition.%d.lag", statsdSafeString(group), statsdSafeString(topic), partition), float64(lag), nil)
}

func (w *StatsDWriter) WriteObservationSummary(duration time.Duration, observationCount int64, brokerCount, topicCount, groupCount, partitionCount int, tags Tags) {
	tagArray := []string{}
	if w.tagFormat == TagFmtDataDog {
		for name, value := range tags {
			tagArray = append(tagArray, fmt.Sprintf("%s:%s", name, value))
		}
	}

	w.gsw.Gauge("observation.observations", float64(observationCount), tagArray)
	w.gsw.Gauge("observation.brokers", float64(brokerCount), tagArray)
	w.gsw.Gauge("observation.consumer_groups", float64(groupCount), tagArray)
	w.gsw.Gauge("observation.topics", float64(topicCount), tagArray)
	w.gsw.Gauge("observation.partitions", float64(partitionCount), tagArray)
	w.gsw.Timing("observation.duration.ms", float64(duration.Nanoseconds()/1000/1000), tagArray)
}

func (w *StatsDWriter) Flush() {
}

func statsdSafeString(value string) string {
	safeString := value
	safeString = statsdReplaceChars.Replace(safeString)
	safeString = statsdRemoveChars.Replace(safeString)
	return statsdBadChars.ReplaceAllLiteralString(safeString, "_")
}
