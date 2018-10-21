package count

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_WorkerPool_ProcessFile(t *testing.T) {
	substring := []byte("SomeSubstring")
	pool, results, tearDown := createWorkerPool(substring)
	defer tearDown()

	numberOfOccurrences := uint64(5)
	t.Run(fmt.Sprintf("When file with %d occurences of substring is processed", numberOfOccurrences), func(t *testing.T) {
		file := createTmpFileFilledWith(t, substring, numberOfOccurrences)

		source, err := NewSource(file.Name(), nil)
		require.NoError(t, err)

		pool.process(source, nil)
		t.Run("It must return correct subtotal", func(t *testing.T) {
			expectSubtotal(t, results, numberOfOccurrences, file.Name(), time.Second)
		})
	})
}

func Test_WorkerPool_ProcessURL(t *testing.T) {
	substring := []byte("SomeSubstring")
	pool, results, tearDown := createWorkerPool(substring)
	defer tearDown()

	numberOfOccurrences := uint64(5)
	t.Run(fmt.Sprintf("When webpage with %d occurences of substring is processed", numberOfOccurrences), func(t *testing.T) {
		s := spawnServer(t, substring, numberOfOccurrences)
		defer s.Close()

		source, err := NewSource(s.URL, &http.Client{Timeout: time.Hour})
		require.NoError(t, err)

		pool.process(source, nil)
		t.Run("It must return correct subtotal", func(t *testing.T) {
			expectSubtotal(t, results, numberOfOccurrences, s.URL, time.Second)
		})
	})
}

func Test_WorkerPool_Stop(t *testing.T) {
	substring := []byte("SomeSubstring")
	stopWorkers := make(chan struct{})
	pool, _, tearDown := createWorkerPool(substring, stopWorkers)
	defer tearDown()

	t.Run("When /dev/urandom is processed and pool.process doesn't block", func(t *testing.T) {
		source, err := NewSource("/dev/urandom", nil)
		require.NoError(t, err)

		pool.process(source, nil)
		t.Run("And stop channel is closed", func(t *testing.T) {
			close(stopWorkers)

			t.Run("It must stop successfully", func(t *testing.T) {
				expectWorkersToStopIn(t, pool.wg, time.Second)
			})
		})
	})
}

func createWorkerPool(substring []byte, optionalStop ...chan struct{}) (*WorkerPool, <-chan *Result, func()) {
	tasks := make(chan *Source)
	results := make(chan *Result)
	stop := make(chan struct{})
	if len(optionalStop) > 0 {
		stop = optionalStop[0]
	}
	worker := workerFunc(results, stop, substring)
	pool := newWorkerPool(tasks, 1, worker, &sync.WaitGroup{})

	return pool, results, func() { close(tasks) }
}

func createTmpFileFilledWith(t *testing.T, data []byte, repeats uint64) *os.File {
	file, err := ioutil.TempFile("", "test")
	require.NoError(t, err)
	writeData(t, file, data, repeats)
	require.NoError(t, file.Close())
	return file
}

func expectSubtotal(t *testing.T, results <-chan *Result, subtotal uint64, origin string, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	select {
	case result := <-results:
		require.Equal(t, origin, result.origin)
		require.Equal(t, big.NewInt(int64(subtotal)), result.subtotal)
	case <-timer.C:
		require.FailNowf(t, "", "No result after %v", timeout)
	}
}

func expectWorkersToStopIn(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-timer.C:
		require.FailNowf(t, "", "Didn't stop after %v", timeout)
	}
}

func spawnServer(t *testing.T, data []byte, repeats uint64) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeData(t, w, data, repeats)
	}))
}

func writeData(t *testing.T, writer io.Writer, data []byte, times uint64) {
	for i := uint64(0); i < times; i++ {
		_, err := writer.Write(data)
		require.NoError(t, err)
	}
}
