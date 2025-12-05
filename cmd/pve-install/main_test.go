package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	assert.Equal(t, "dev", version)
	assert.Equal(t, "none", commit)
	assert.Equal(t, "unknown", date)
}
