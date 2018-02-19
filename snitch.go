package main

import (
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	units "github.com/docker/go-units"
)

type Snitch struct {
	observer *Observer
	config   *ObserveConfig
	client   sarama.Client
	doneCh   chan bool
	termCh   chan bool
}

func NewSnitch(observer *Observer, config *ObserveConfig) *Snitch {
	return &Snitch{observer: observer, config: config}
}

type TopicSet map[string]map[int32]int64

func (s *Snitch) Connect(brokers []string) chan bool {
	config := sarama.NewConfig()
	config.ClientID = "kafka-snitch"
	config.Version = sarama.V0_9_0_0

	readyCh := make(chan bool)
	retryTimeout := 10 * time.Second
	go func() {
		for {
			var err error
			if s.client, err = sarama.NewClient(brokers, config); err == nil {
				break
			}

			log.Debugf("Problem connecting to Kafka: %v", err)
			log.Infof("Couldn't connect to Kafka! Trying again in %v", retryTimeout)
			time.Sleep(retryTimeout)
		}

		log.Infof("Connected to the Kafka cluster!")
		s.doneCh = make(chan bool)
		s.termCh = make(chan bool)
		readyCh <- true
	}()

	return readyCh
}

func (s *Snitch) Run() {
	tally := NewTally()
	for {
		select {
		case <-time.After(10 * time.Second):
			log.Info("Beginning observation")
			observationStart := time.Now()

			s.observe(tally)

			observationDuration := time.Since(observationStart)
			s.observer.Observation(observationDuration,
				tally.BrokerCount(), tally.TopicCount(), tally.GroupCount(), tally.PartitionCount())

			log.WithFields(log.Fields{
				"brokers":     tally.BrokerCount(),
				"topics":      tally.TopicCount(),
				"groups":      tally.GroupCount(),
				"partitions":  tally.PartitionCount(),
				"duration_ms": observationDuration.Nanoseconds() / 1000 / 1000,
			}).Infof("Observation complete in %v", strings.ToLower(units.HumanDuration(observationDuration)))

			tally.Reset()
			break

		case <-s.doneCh:
			s.client.Close()
			log.Info("Disconnected from the Kafka cluster")
			s.termCh <- true
			break
		}
	}
}

func (s *Snitch) observe(tally *Tally) {
	log.Debug("Refreshing topics.")

	topicSet := make(TopicSet)
	s.client.RefreshMetadata()

	topics, err := s.client.Topics()
	if err != nil {
		log.Errorf("Problem refreshing topics! %v", err.Error())
		return
	}

	for _, topic := range topics {
		// Don't include internal topics
		if strings.HasPrefix(topic, "__") {
			continue
		}

		partitions, err := s.client.Partitions(topic)
		if err != nil {
			log.Errorf("Problem fetching partitions! %v", err.Error())
			continue
		}

		topicSet[topic] = make(map[int32]int64)
		for _, partition := range partitions {
			offset, err := s.client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				log.WithFields(log.Fields{
					"topic":     topic,
					"partition": partition,
				}).Errorf("Problem fetching offset! %v", err)
				continue
			}

			topicSet[topic][partition] = offset
		}
	}

	log.Debugf("Refreshed %d topics", len(topicSet))

	var wg sync.WaitGroup

	// Lookup group data using the metadata
	for _, broker := range s.client.Brokers() {
		if !s.canObserveBroker(broker) {
			continue
		}

		broker.Open(s.client.Config())
		if _, err := broker.Connected(); err != nil {
			log.WithFields(log.Fields{
				"broker": broker.ID(),
				"addr":   broker.Addr(),
			}).Errorf("Could not connect to broker: %v", err)
			continue
		}

		log.WithFields(log.Fields{
			"broker": broker.ID(),
			"addr":   broker.Addr(),
		}).Debugf("Connected to broker %s", broker.ID())

		wg.Add(1)

		go func(broker *sarama.Broker) {
			defer wg.Done()
			s.observeBroker(tally, broker, topicSet)
		}(broker)
	}

	wg.Wait()
}

func (s *Snitch) observeBroker(tally *Tally, broker *sarama.Broker, topicSet TopicSet) {
	groupsRequest := new(sarama.ListGroupsRequest)
	groupsResponse, err := broker.ListGroups(groupsRequest)

	if err != nil {
		log.WithFields(log.Fields{
			"broker": broker.ID(),
		}).Errorf("Could not list groups: %v", err)
		return
	}

	for group, protocolType := range groupsResponse.Groups {
		log.WithFields(log.Fields{
			"group":        group,
			"protocolType": protocolType,
		}).Debug("Found group")

		for topic, data := range topicSet {
			var totalLag int64

			offsetsRequest := new(sarama.OffsetFetchRequest)
			offsetsRequest.Version = 1
			offsetsRequest.ConsumerGroup = group
			for partition := range data {
				offsetsRequest.AddPartition(topic, partition)
			}

			offsetsResponse, err := broker.FetchOffset(offsetsRequest)
			if err != nil {
				log.WithFields(log.Fields{
					"consumer_group": group,
					"broker":         broker.ID(),
				}).Errorf("Could not get offsets: %v", err)

				continue
			}

			for _, blocks := range offsetsResponse.Blocks {
				for partition, block := range blocks {
					// Block offset is -1 if the group isn't active on the topic
					if block.Offset < 0 {
						continue
					}

					logEndOffset := data[partition]
					lag := logEndOffset - block.Offset
					if lag < 0 {
						lag = 0
					}
					totalLag += lag

					log.WithFields(log.Fields{
						"consumer_group": group,
						"topic":          topic,
						"partition":      partition,
						"lag":            lag,
						"broker":         broker.ID(),
					}).Debugf("Observed %s consumer group on topic %s, partition %d", group, topic, partition)

					tally.Add(broker.ID(), topic, group, partition)
					s.observer.PartitionLag(group, topic, partition, logEndOffset, block.Offset, lag)
				}
			}

			s.observer.TopicLag(group, topic, totalLag)
		}
	}
}

func (s *Snitch) canObserveBroker(broker *sarama.Broker) bool {
	if len(s.config.Brokers) == 0 {
		return true
	}

	test := int(broker.ID())
	for _, id := range s.config.Brokers {
		if test == id {
			return true
		}
	}

	return false
}

func (s *Snitch) Close() {
	s.doneCh <- true
	<-s.termCh
}
