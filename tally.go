package main

type Tally struct {
	brokers    map[int32]int
	topics     map[string]int
	partitions map[string]map[int32]int
	groups     map[string]int
}

type brokerBucket map[int32]int
type topicBucket map[string]int
type groupBucket map[string]int
type partitionBucket map[string]map[int32]int

func NewTally() *Tally {
	return &Tally{
		brokers:    make(brokerBucket),
		topics:     make(topicBucket),
		groups:     make(groupBucket),
		partitions: make(partitionBucket),
	}
}

func (t *Tally) Add(broker int32, topic, group string, partition int32) {
	t.brokers[broker] = 1
	t.groups[group] = 1
	t.topics[topic] = 1

	partitionMap := t.partitions[topic+group]
	if partitionMap == nil {
		partitionMap = make(map[int32]int)
		t.partitions[topic+group] = partitionMap
	}

	partitionMap[partition] = 1
}

func (t *Tally) BrokerCount() (count int) {
	for _, v := range t.brokers {
		count = count + v
	}

	return
}

func (t *Tally) GroupCount() (count int) {
	for _, v := range t.groups {
		count = count + v
	}

	return
}

func (t *Tally) TopicCount() (count int) {
	for _, v := range t.topics {
		count = count + v
	}

	return
}

func (t *Tally) PartitionCount() (count int) {
	for _, partitionMap := range t.partitions {
		for _, partitionFlag := range partitionMap {
			count = count + partitionFlag
		}
	}

	return
}

func (t *Tally) Reset() {
	for broker, _ := range t.brokers {
		t.brokers[broker] = 0
	}

	for group, _ := range t.groups {
		t.groups[group] = 0
	}

	for topic, _ := range t.topics {
		t.topics[topic] = 0
	}

	for _, partitionMap := range t.partitions {
		for partition, _ := range partitionMap {
			partitionMap[partition] = 0
		}
	}
}
