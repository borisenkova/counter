package count

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_WorkerPool_ProcessFile(t *testing.T) {
	substring := []byte("SomeSubstring")
	pool, tasks := createWorkerPool(context.Background(), substring)

	numberOfOccurrences := 5
	t.Run(fmt.Sprintf("When file with %d occurences of substring is processed", numberOfOccurrences), func(t *testing.T) {
		file := createTmpFileFilledWith(t, substring, numberOfOccurrences)

		source, err := newSource(file.Name(), time.Minute)
		require.NoError(t, err)
		tasks <- source
		t.Run("It must return correct subtotal", func(t *testing.T) {
			expectSubtotal(t, pool.results, numberOfOccurrences, file.Name(), time.Second)
		})
	})
}

func Test_WorkerPool_ProcessURL(t *testing.T) {
	substring := []byte("SomeSubstring")
	pool, tasks := createWorkerPool(context.Background(), substring)

	numberOfOccurrences := 5
	t.Run(fmt.Sprintf("When webpage with %d occurences of substring is processed", numberOfOccurrences), func(t *testing.T) {
		s := spawnServer(t, substring, numberOfOccurrences)
		defer s.Close()

		source, err := newSource(s.URL, time.Minute)
		require.NoError(t, err)

		tasks <- source
		t.Run("It must return correct subtotal", func(t *testing.T) {
			expectSubtotal(t, pool.results, numberOfOccurrences, s.URL, time.Second)
		})
	})
}

func Test_WorkerPool_Stop(t *testing.T) {
	substring := []byte("SomeSubstring")
	ctx, cancel := context.WithCancel(context.Background())
	s := spawnSlowServer()
	defer s.Close()

	pool, _ := createWorkerPool(ctx, substring)

	t.Run("When really slow origin is processed", func(t *testing.T) {
		source, err := newSource(s.URL, time.Minute)
		require.NoError(t, err)

		pool.process(source)
		t.Run("And context is canceled", func(t *testing.T) {
			time.Sleep(time.Second)
			cancel()

			t.Run("It must stop successfully", func(t *testing.T) {
				expectWorkersToStopIn(t, pool, time.Second)
			})
		})
	})
}

func Test_WorkerPool_ShutdownWorkerIfInactive(t *testing.T) {
	inactivityTimeout := time.Second
	worker := workerFunc([]byte("1"), inactivityTimeout)
	pool := newWorkerPool(context.Background(), 1, worker)
	tasks := make(chan source, 1)
	pool.consume(tasks)

	t.Run("When one task is processed", func(t *testing.T) {
		source, err := newSource("http://localhost", inactivityTimeout)
		require.NoError(t, err)
		tasks <- source
		<-pool.results

		t.Run("One worker must still be running", func(t *testing.T) {
			require.Equal(t, 1, pool.getNumberOfWorkers())

			t.Run("Then if no tasks is provided during some period of time", func(t *testing.T) {
				timer := time.NewTimer(2 * inactivityTimeout)
				<-timer.C

				t.Run("It must stop running worker", func(t *testing.T) {
					require.Equal(t, 0, pool.getNumberOfWorkers())
				})
			})
		})
	})
}

func createWorkerPool(ctx context.Context, substring []byte) (*workerPool, chan source) {
	worker := workerFunc(substring, time.Minute)
	pool := newWorkerPool(ctx, 1, worker)
	tasks := make(chan source, 1)
	pool.consume(tasks)

	return pool, tasks
}

func createTmpFileFilledWith(t *testing.T, data []byte, repeats int) *os.File {
	file, err := ioutil.TempFile("", "test")
	require.NoError(t, err)
	writeData(t, file, data, repeats)
	require.NoError(t, file.Close())
	return file
}

func expectSubtotal(t *testing.T, results <-chan *result, subtotal int, origin string, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	select {
	case result := <-results:
		require.Equal(t, origin, result.origin)
		require.Equal(t, big.NewInt(int64(subtotal)), result.subtotal)
	case <-timer.C:
		require.FailNowf(t, "", "No result after %v", timeout)
	}
}

func expectWorkersToStopIn(t *testing.T, pool *workerPool, timeout time.Duration) {
	done := make(chan struct{})

	go func() {
		pool.wg.Wait()
		done <- struct{}{}
	}()

	expectStopIn(t, done, timeout)
}

func expectStopIn(t *testing.T, done <-chan struct{}, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	select {
	case <-done:
	case <-timer.C:
		require.FailNowf(t, "", "Didn't stop after %v", timeout)
	}
}

func spawnServer(t *testing.T, data []byte, repeats int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeData(t, w, data, repeats)
	}))
}

func spawnSlowServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(time.Minute):
		case <-r.Context().Done():
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func writeData(t *testing.T, writer io.Writer, data []byte, times int) {
	for i := 0; i < times; i++ {
		_, err := writer.Write(data)
		require.NoError(t, err)
	}
}
