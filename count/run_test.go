package count

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func Test_Run_Stop(t *testing.T) {
	substring := []byte("SomeSubstring")
	ctx := context.Background()
	s := spawnServer(t, substring, 1)
	defer s.Close()

	t.Run("When URL is processed", func(t *testing.T) {
		done := make(chan struct{})
		go func() {
			Run(ctx, bytes.NewBufferString(s.URL), substring, 1, time.Minute, time.Minute)
			done <- struct{}{}
		}()

		t.Run("It must stop successfully after processing", func(t *testing.T) {
			expectStopIn(t, done, time.Second)
		})
	})
}

func Test_Run_Stop_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := spawnSlowServer()
	defer s.Close()

	t.Run("When really slow origin is processed", func(t *testing.T) {
		done := make(chan struct{})
		go func() {
			Run(ctx, bytes.NewBufferString(s.URL), []byte("SomeSubstring"), 1, time.Minute, time.Minute)
			done <- struct{}{}
		}()

		t.Run("And context is canceled", func(t *testing.T) {
			cancel()

			t.Run("It must stop successfully", func(t *testing.T) {
				expectStopIn(t, done, time.Second)
			})
		})
	})
}
