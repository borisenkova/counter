package count

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Run_Stop(t *testing.T) {
	tasks := make(chan *Source)
	results := make(chan *Result)
	substring := []byte("SomeSubstring")

	t.Run("When really long input is processed", func(t *testing.T) {
		stop := make(chan struct{})
		done := make(chan struct{})
		input := bytes.NewBufferString("/dev/urandom\n/dev/urandom\n/dev/urandom")
		go func() {
			Run(input, substring, 2, tasks, results, stop, nil)
			done <- struct{}{}
		}()
		t.Run("And stop channel is closed", func(t *testing.T) {
			time.Sleep(time.Second)
			close(stop)

			t.Run("It must successfully", func(t *testing.T) {
				timeout := time.Second
				timer := time.NewTimer(timeout)
				select {
				case <-done:
				case <-timer.C:
					require.FailNowf(t, "", "Didn't stop after %v", timeout)
				}
			})
		})
	})
}
