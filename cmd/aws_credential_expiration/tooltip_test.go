package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcat_joinsItemsCorrectly(t *testing.T) {
	actual := concat(map[string]time.Time{"one": time.Now(), "two": time.Now(), "four": time.Now()})
	assert.Equal(t, "four, one and two", actual)
}
