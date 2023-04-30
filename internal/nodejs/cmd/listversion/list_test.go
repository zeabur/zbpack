package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNodeVersionsList(t *testing.T) {
	v, err := getNodeVersionsList()
	assert.NoError(t, err)
	assert.NotEmpty(t, v)

	for i := range v {
		// ignore the first entry
		if i == 0 {
			continue
		}

		assert.True(t, v[i-1].Major() >= v[i].Major())
	}
}
