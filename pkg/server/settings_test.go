package server

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	dialer := NewMockDialer(ctrl)
	dialer.EXPECT().String().Return("test").AnyTimes()

	middleware := NewMockMiddleware(ctrl)
	middleware.EXPECT().String().Return("test").AnyTimes()

	testCases := map[string]struct {
		settings Settings
		s        string
	}{
		"empty_settings": {
			settings: Settings{
				ListeningAddress: ptrTo("localhost:53"),
				Dialer:           dialer,
			},
			s: `Server settings:
├── Listening address: localhost:53
└── Upstream resolver connection type: test`,
		},
		"non_empty_settings": {
			settings: Settings{
				ListeningAddress: ptrTo(":8000"),
				Dialer:           dialer,
				Middlewares:      []Middleware{middleware},
			},
			s: `Server settings:
├── Listening address: :8000
├── Upstream resolver connection type: test
└── Middlewares:
    └── test`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := testCase.settings.String()

			assert.Equal(t, testCase.s, s)
		})
	}
}
