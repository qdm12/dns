package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_SetDefaults(t *testing.T) {
	settings := Settings{}
	settings.SetDefaults()

	assert.Greater(t, settings.MaxEntries, 1)
	assert.Greater(t, settings.TTL, time.Second)
}
