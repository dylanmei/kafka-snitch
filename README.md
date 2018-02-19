kafka-snitch
------------

A simple application to observe interesting details from a [Kafka](http://kafka.apache.org) 0.9+ cluster, including consumer group lag.

`kafka-snitch` observes these details:

- [x] consumer group lag
- [ ] empty topics

`kafka-snitch` supports these reporters:

- [x] logger
- [x] influxdb
- [-] statsd (in progress)
- [ ] prometheus

# metrics

Observations

- kafka.snitch.observation.brokers
- kafka.snitch.observation.topics
- kafka.snitch.observation.consumer_groups
- kafka.snitch.observation.partitions
- kafka.snitch.observation.duration.ms

Consumers

- kafka.snitch.consumers.<consumer_group>.topic.<topic>.lag
- kafka.snitch.consumers.<consumer_group>.topic.<topic>.partition.<partition>.log_end_offset
- kafka.snitch.consumers.<consumer_group>.topic.<topic>.partition.<partition>.consumer_offset
- kafka.snitch.consumers.<consumer_group>.topic.<topic>.partition.<partition>.lag

# example

Build and run `kafka-snitch`

```
make
bin/kafka-snitch -brokers=localhost:9092
```

Show the help

```
$ kafka-snitch --help
Usage of bin/kafka-snitch:
  -brokers string
    	The hostname:port of one or more Kafka brokers
  -influxdb.buffer-size int
    	The maximum number of points to buffer before flushing to InfluxDB (default 1000)
  -influxdb.database string
    	The target InfluxDB database name
  -influxdb.flush-interval int
    	The number of seconds to wait before flushing to InfluxDB (default 60)
  -influxdb.http.url string
    	The http://hostname:port of an InfluxDB HTTP endpoint
  -influxdb.precision string
    	The precision of points written to InfluxDB: "s", "ms", "us" (default "us")
  -influxdb.retention-policy string
    	The target InfluxDB database retention policy name
  -influxdb.udp.addr string
    	The hostname:port of an InfluxDB UDP endpoint
  -log.format string
    	Logging format: text, json (default "text")
  -log.level string
    	Logging level: debug, info, warning, error (default "info")
  -observe.broker value
    	A broker-id to include when observing offsets; other brokers will be ignored
  -run.once
    	Whether to run-and-exit, or run continously
  -run.snooze duration
    	The amount of time to sleep between observations (default 10s)
  -statsd.addr string
    	The hostname:port of a StatsD UDP endpoint
  -statsd.tagfmt string
    	The tagging-format of the metric payloads: none, DataDog (default "none")
  -version
    	Print the current version
```

Report to InfluxDB and print out debug logs in JSON format, but only for brokers 1:

```
make
bin/kafka-snitch -brokers=localhost:9092 \
  -influxdb.http.url=http://localhost:8081 \
  -log.level=debug -log.format=json \
  -observe.broker=1
```

## notes

Applying an `-observe.broker` constraint comes in handy when running _snitch_ on each host alongside the broker process.

There's a Docker image here: https://hub.docker.com/r/dylanmei/kafka-snitch
