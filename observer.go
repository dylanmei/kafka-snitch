package main

import (
	"time"
)

type Tags map[string]string

type Observer struct {
	writers          []ObserverWriter
	observationCount int64
}

type ObserverWriter interface {
	WriteConsumerTopicLag(group, topic string, lag int64, tags Tags)
	WriteConsumerPartitionLag(group, topic string, partition int, logEndOffset, consumerOffset, lag int64, tags Tags)
	WriteObservationSummary(duration time.Duration, observationCount int64, brokerCount, topicCount, groupCount, partitionCount int, tags Tags)
	Flush()
}

func NewObserver() *Observer {
	return &Observer{
		writers: []ObserverWriter{},
	}
}

func (o *Observer) AddWriter(w ObserverWriter) {
	o.writers = append(o.writers, w)
}

func (o *Observer) Flush() {
	for _, w := range o.writers {
		w.Flush()
	}
}

func (o *Observer) TopicLag(group, topic string, lag int64) {
	for _, w := range o.writers {
		w.WriteConsumerTopicLag(group, topic, lag, Tags{})
	}
}

func (o *Observer) PartitionLag(group, topic string, partition int32, logEndOffset, consumerOffset, lag int64) {
	for _, w := range o.writers {
		w.WriteConsumerPartitionLag(group, topic, int(partition), logEndOffset, consumerOffset, lag, Tags{})
	}
}

func (o *Observer) Observation(duration time.Duration, brokerCount, topicCount, groupCount, partitionCount int) {
	o.observationCount++

	for _, w := range o.writers {
		w.WriteObservationSummary(duration, o.observationCount, brokerCount, topicCount, groupCount, partitionCount, Tags{})
	}
}
