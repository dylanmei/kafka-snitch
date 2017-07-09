package main

import (
	"strconv"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

type Observer struct {
	writer           *InfluxDBWriter
	observationCount int64
}

func (o *Observer) ConsumerGroup(topic, group string, partition int32, logSize, offset, lag int64) {
	tags := map[string]string{
		"topic":     topic,
		"group":     group,
		"partition": strconv.Itoa(int(partition)),
	}

	fields := map[string]interface{}{
		"log_size": logSize,
		"offset":   offset,
		"lag":      lag,
	}

	point, _ := influxdb.NewPoint("kafka_snitch_consumer_group", tags, fields, time.Now())
	o.writer.Write(point)
}

func (o *Observer) Observation(duration time.Duration, brokerCount, topicCount, groupCount, partitionCount int) {
	tags := map[string]string{}

	o.observationCount++
	fields := map[string]interface{}{
		"observation_count": o.observationCount,
		"duration":          duration.Nanoseconds(),
		"broker_count":      brokerCount,
		"topic_count":       topicCount,
		"group_count":       groupCount,
		"partition_count":   partitionCount,
	}

	point, _ := influxdb.NewPoint("kafka_snitch_observation", tags, fields, time.Now())
	o.writer.Write(point)
}

func (o *Observer) Close() {
	o.writer.Close()
}
