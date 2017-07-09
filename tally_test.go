package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_tally(t *testing.T) {
	tally := NewTally()
	tally.Add(1, "test", "foo", 1)
	tally.Add(2, "test", "bar", 1)
	tally.Add(2, "test", "bar", 2)

	assert.Equal(t, 2, tally.BrokerCount())
	assert.Equal(t, 1, tally.TopicCount())
	assert.Equal(t, 2, tally.GroupCount())
	assert.Equal(t, 3, tally.PartitionCount())

	tally.Reset()

	assert.Equal(t, 0, tally.BrokerCount())
	assert.Equal(t, 0, tally.TopicCount())
	assert.Equal(t, 0, tally.GroupCount())
	assert.Equal(t, 0, tally.PartitionCount())
}
