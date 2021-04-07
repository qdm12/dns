package splash

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AnnouncementExpiration(t *testing.T) {
	t.Parallel()
	if len(announcementExpiration) == 0 {
		return
	}
	_, err := time.Parse("2006-01-02", announcementExpiration)
	assert.NoError(t, err)
}
