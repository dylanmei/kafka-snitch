kafka-snitch
------------

A simple application to extract interesting details from a [Kafka](http://kafka.apache.org) 0.9+ cluster, including consumer group lag.

`kafka-snitch` supports these reporters:

- logger
- influxdb

# example

Build and run `kafka-snitch`

```
make
bin/kafka-snitch -brokers=localhost:9092
```

## notes

We're currently using [dep](https://github.com/golang/dep) for vendoring.
