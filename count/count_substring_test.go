package count

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountSubstring(t *testing.T) {
	expectCount(t, "aaaaaaaa", 10, "a", 8)
	expectCount(t, "abcdabcd", 3, "abcd", 2)
	expectCount(t, "ababababababa", 4, "aba", 3)
}

func expectCount(t *testing.T, source string, bufLen int, substr string, expectedCount int) {
	sourceBytes := ioutil.NopCloser(bytes.NewBuffer([]byte(source)))
	buf := make([]byte, bufLen)
	substrBytes := []byte(substr)
	count, err := countSubstr(sourceBytes, buf, substrBytes)
	require.NoError(t, err)
	require.Equal(t, int64(expectedCount), count.Int64())
}
