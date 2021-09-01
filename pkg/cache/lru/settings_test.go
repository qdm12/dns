package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_setDefaults(t *testing.T) {
	settings := Settings{}
	settings.setDefaults()

	assert.Greater(t, settings.MaxEntries, 1)
	assert.NotNil(t, settings.Metrics)
}
