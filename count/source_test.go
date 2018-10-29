package count

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_isHTTPURL(t *testing.T) {
	require.True(t, isHTTPURL("https://some.url"), "URL with https scheme is supported")
	require.True(t, isHTTPURL("http://some.url"), "URL with http scheme is supported")
	require.False(t, isHTTPURL("some.url"), "URL without scheme is not supported")
}

func Test_Source_Get_NotRegularFile(t *testing.T) {
	t.Run("When /dev/urandom is provided as origin", func(t *testing.T) {
		source, err := newSource("/dev/urandom", time.Minute)
		t.Run("It must return error indicating that source is unknown", func(t *testing.T) {
			require.Error(t, err)
			require.Equal(t, errUnknownSourceStr, err.Error())
			require.Nil(t, source)
		})
	})
}
