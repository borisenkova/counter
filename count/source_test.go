package count

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isHTTPURL(t *testing.T) {
	require.True(t, isHTTPURL("https://some.url"), "URL with https scheme is supported")
	require.True(t, isHTTPURL("http://some.url"), "URL with http scheme is supported")
	require.False(t, isHTTPURL("some.url"), "URL without scheme is not supported")
}
