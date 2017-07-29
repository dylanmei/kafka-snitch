kafka-snitch
------------

A simple application to observe interesting details from a [Kafka](http://kafka.apache.org) 0.9+ cluster, including consumer group lag.

`kafka-snitch` observes these details:

- [x] consumer group lag
- [ ] empty topics 

`kafka-snitch` supports these reporters:

- [x] logger
- [x] influxdb
- [ ] prometheus

# example

Build and run `kafka-snitch`

```
make
bin/kafka-snitch -brokers=localhost:9092
```

Report to InfluxDB and print out debug logs in JSON format, but only for brokers 2 and 4:

```
make
bin/kafka-snitch -brokers=localhost:9092 \
  -influxdb.http.url=http://localhost:8081 \
  -log.level=debug -log.format=json \
  -observe.broker=2 -observe.broker=4
```

That's a contrived example, but `-observe.broker` comes in handy (speed and simplicity) when running these as sidecar containers or on each host alongside a broker process.

## notes

There's a Docker container here: https://hub.docker.com/r/dylanmei/kafka-snitch

We're currently using [dep](https://github.com/golang/dep) for vendoring.