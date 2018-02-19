package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

type SnitchConfig struct {
	Brokers   string
	LogLevel  string
	LogFormat string
	InfluxDB  InfluxDBConfig
	Observe   ObserveConfig
}

const LogFormatText = "text"
const LogFormatJSON = "json"

type InfluxDBConfig struct {
	Database        string
	RetentionPolicy string
	Precision       string
	HTTPConfig      influxdb.HTTPConfig
	UDPConfig       influxdb.UDPConfig
	BufferSize      int
	FlushInterval   int
}

type ObserveConfig struct {
	Brokers IDArray
}

type IDArray []int

func (a *IDArray) String() string {
	return "not implemented"
}

func (a *IDArray) Set(value string) error {
	tmp, err := strconv.Atoi(value)
	if err == nil {
		*a = append(*a, tmp)
	}

	return nil
}

func (config *SnitchConfig) Parse() {
	flag.StringVar(&config.Brokers,
		"brokers", "", "The hostname:port of one or more Kafka brokers")

	flag.Var(&config.Observe.Brokers,
		"observe.broker", "A broker-id to include when observing offsets; other brokers will be ignored")

	flag.StringVar(&config.InfluxDB.UDPConfig.Addr,
		"influxdb.udp.addr", "", "The hostname:port of an InfluxDB UDP endpoint")
	flag.StringVar(&config.InfluxDB.HTTPConfig.Addr,
		"influxdb.http.url", "", "The http://hostname:port of an InfluxDB HTTP endpoint")
	flag.IntVar(&config.InfluxDB.BufferSize,
		"influxdb.buffer-size", 1000, "The maximum number of points to buffer before flushing to InfluxDB")
	flag.IntVar(&config.InfluxDB.FlushInterval,
		"influxdb.flush-interval", 60, "The number of seconds to wait before flushing to InfluxDB")
	flag.StringVar(&config.InfluxDB.Database,
		"influxdb.database", "", "The target InfluxDB database name")
	flag.StringVar(&config.InfluxDB.RetentionPolicy,
		"influxdb.retention-policy", "", "The target InfluxDB database retention policy name")
	flag.StringVar(&config.InfluxDB.Precision,
		"influxdb.precision", "us", "The precision of points written to InfluxDB: \"s\", \"ms\", \"us\"")

	flag.StringVar(&config.LogLevel, "log.level", log.InfoLevel.String(), "Logging level: debug, info, warning, error")
	flag.StringVar(&config.LogFormat, "log.format", LogFormatText, "Logging format: text, json")

	showVersion := flag.Bool("version", false, "Print the current version")

	flag.Parse()
	if *showVersion {
		PrintVersion(os.Stdout)
		os.Exit(0)
	}

	SetLogFormat(config.LogFormat)
	SetLogLevel(config.LogLevel)
}

func SetLogFormat(f string) {
	if f == LogFormatJSON {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}
}

func SetLogLevel(l string) {
	level, err := log.ParseLevel(l)
	if err != nil {
		log.Fatalf("Oops! %v", err)
	}

	log.SetLevel(level)
}
