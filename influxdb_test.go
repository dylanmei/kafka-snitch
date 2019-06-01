package main

import (
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

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

func (d DummyMetricsClient) QueryAsChunk(q influxdb.Query) (*influxdb.ChunkedResponse, error) {
	return &influxdb.ChunkedResponse{}, nil
}

func (d DummyMetricsClient) Ping(t time.Duration) (time.Duration, string, error) {
	return 0 * time.Second, "", nil
}

func Test_metrics_writer_writes_a_point(t *testing.T) {
	batchCh := make(chan influxdb.BatchPoints)
	w, _ := newInfluxDBWriter(&InfluxDBConfig{BufferSize: 0}, DummyMetricsClient{
		writeFunc: func(bp influxdb.BatchPoints) error {
			batchCh <- bp
			return nil
		},
	})

	defer w.Close()
	w.Write(createPoints(1)[0])

	var written bool

	select {
	case <-time.After(10 * time.Millisecond):
		break
	case <-batchCh:
		written = true
	}

	if !written {
		t.Error("Expected a point to be written")
	}
}

func Test_metrics_writer_flushes_points(t *testing.T) {
	numPoints := 10
	points := make([]*influxdb.Point, 0)

	w, _ := newInfluxDBWriter(&InfluxDBConfig{BufferSize: 100}, DummyMetricsClient{
		writeFunc: func(bp influxdb.BatchPoints) error {
			for _, p := range bp.Points() {
				points = append(points, p)
			}
			return nil
		},
	})
	defer w.Close()

	for _, p := range createPoints(numPoints) {
		w.Write(p)
	}

	w.Flush()

	if len(points) != numPoints {
		t.Errorf("Expecting %d points after flushing points, but got: %d", numPoints, len(points))
	}
}

func Test_metrics_writer_flushs_points_when_the_buffer_size_is_reached(t *testing.T) {
	bufferSize := 5
	points := make([]*influxdb.Point, 0)
	writtenCh := make(chan bool)

	w, _ := newInfluxDBWriter(&InfluxDBConfig{BufferSize: bufferSize}, DummyMetricsClient{
		writeFunc: func(bp influxdb.BatchPoints) error {
			for _, p := range bp.Points() {
				points = append(points, p)
			}
			writtenCh <- true
			return nil
		},
	})
	defer w.Close()

	for _, p := range createPoints(bufferSize) {
		w.Write(p)
	}

	<-writtenCh
	count := len(points)

	if count != bufferSize {
		t.Errorf("Expecting %d points after reaching buffer size, but got: %d", bufferSize, count)
	}
}

func Test_metrics_writer_flushed_points_when_closing(t *testing.T) {
	var written bool
	w, _ := newInfluxDBWriter(&InfluxDBConfig{BufferSize: 10}, DummyMetricsClient{
		writeFunc: func(bp influxdb.BatchPoints) error {
			written = true
			return nil
		},
	})

	w.Write(createPoints(1)[0])
	w.Close()

	if !written {
		t.Error("Expected points to be written when closing the writer")
	}
}

func createPoints(num int) []*influxdb.Point {
	points := make([]*influxdb.Point, 0)

	for i := 0; i < num; i++ {
		fields := map[string]interface{}{
			"pizza": "party_" + string(num),
		}

		point, _ := influxdb.NewPoint("foo", nil, fields)
		points = append(points, point)
	}

	return points
}
