package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_safe_statsd_metric_names(t *testing.T) {
	assert.Equal(t, "foo", statsdSafeString("foo"))
	assert.Equal(t, "foo-bar", statsdSafeString("foo@bar"))
	assert.Equal(t, "foo-bar", statsdSafeString(`"foo" 'bar'`))
	assert.Equal(t, "foo_bar", statsdSafeString("foo.bar"))
}
