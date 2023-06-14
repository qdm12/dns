package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	settings := Settings{}
	settings.SetDefaults()

	assert.Greater(t, settings.MaxEntries, uint(1))
	assert.NotNil(t, settings.Metrics)
}
