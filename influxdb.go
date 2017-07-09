package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

type InfluxDBWriter struct {
	config   *InfluxDBConfig
	client   influxdb.Client
	pointsCh chan *influxdb.Point
	flushCh  chan chan bool
	closeCh  chan bool

	bufferSize    int
	bufferTimeout time.Duration
}

const defaultBufferSize = 100

func NewInfluxDBWriter(config *InfluxDBConfig) (*InfluxDBWriter, error) {
	client, err := newInfluxdbClient(config)
	if err != nil {
		return nil, fmt.Errorf("Unable to create InfluxDB client: %v", err)
	}
	return newInfluxDBWriter(config, client)
}

func emptyInfluxDBWriter() *InfluxDBWriter {
	w, _ := newInfluxDBWriter(&InfluxDBConfig{}, DummyMetricsClient{})
	return w
}

func newInfluxDBWriter(config *InfluxDBConfig, client influxdb.Client) (*InfluxDBWriter, error) {
	bufferSize := config.BufferSize
	if bufferSize < 0 {
		bufferSize = 0
	}

	flushInterval := config.FlushInterval
	if flushInterval < 1 {
		flushInterval = 1
	}

	w := &InfluxDBWriter{
		config:        config,
		client:        client,
		pointsCh:      make(chan *influxdb.Point),
		flushCh:       make(chan chan bool),
		closeCh:       make(chan bool),
		bufferSize:    bufferSize,
		bufferTimeout: time.Duration(flushInterval) * time.Second,
	}

	go w.capturePoints()
	return w, nil
}

func newInfluxdbClient(config *InfluxDBConfig) (influxdb.Client, error) {
	if config.HTTPConfig.Addr != "" {
		return influxdb.NewHTTPClient(config.HTTPConfig)
	}

	if config.UDPConfig.Addr != "" {
		return influxdb.NewUDPClient(config.UDPConfig)
	}

	return DummyMetricsClient{}, nil
}

func (w *InfluxDBWriter) Write(point *influxdb.Point) {
	w.pointsCh <- point
}

func (w *InfluxDBWriter) Flush() {
	done := make(chan bool)
	w.flushCh <- done
	<-done
}

func (w *InfluxDBWriter) Close() {
	w.Flush()
	w.closeCh <- true
	w.client.Close()
}

func (w *InfluxDBWriter) flushPoints(points []*influxdb.Point) {
	if len(points) == 0 {
		return
	}

	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:        w.config.Database,
		Precision:       w.config.Precision,
		RetentionPolicy: w.config.RetentionPolicy,
	})

	if err != nil {
		log.Error("Problem creating batch point! %v", err)
		return
	}

	for _, pt := range points {
		bp.AddPoint(pt)
	}

	w.client.Write(bp)
}

func (w *InfluxDBWriter) capturePoints() {
	points := make([]*influxdb.Point, 0)
	timer := time.NewTimer(w.bufferTimeout)

	for {
		select {

		case p := <-w.pointsCh:
			points = append(points, p)

			if w.bufferSize <= len(points) {

				w.flushPoints(points)
				points = make([]*influxdb.Point, 0)

				timer.Reset(w.bufferTimeout)
			}

		case <-timer.C:
			if len(points) > 0 {

				w.flushPoints(points)
				points = make([]*influxdb.Point, 0)
			}

			timer.Reset(w.bufferTimeout)

		case flushed := <-w.flushCh:

			w.flushPoints(points)
			points = make([]*influxdb.Point, 0)

			flushed <- true
			timer.Reset(w.bufferTimeout)

		case <-w.closeCh:
			timer.Stop()
			break
		}
	}
}

type DummyMetricsClient struct {
	writeFunc func(influxdb.BatchPoints) error
	closeFunc func() error
}

func (d DummyMetricsClient) Write(bp influxdb.BatchPoints) error {
	if d.writeFunc != nil {
		return d.writeFunc(bp)
	}
	return nil
}

func (d DummyMetricsClient) Close() error {
	if d.closeFunc != nil {
		return d.closeFunc()
	}
	return nil
}

func (d DummyMetricsClient) Query(q influxdb.Query) (*influxdb.Response, error) {
	return &influxdb.Response{}, nil
}

func (d DummyMetricsClient) Ping(t time.Duration) (time.Duration, string, error) {
	return 0 * time.Second, "", nil
}
