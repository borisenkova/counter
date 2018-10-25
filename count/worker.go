package count

import (
	"bytes"
	"errors"
	"io"
	"log"
	"math/big"
	"sync"
)

const averageWebpageSize = 2e+6
const errStopSignalStr = "got.stop.signal"

func countSubstr(source io.ReadCloser, buf, substr []byte) (total *big.Int, err error) {
	defer source.Close()
	total = big.NewInt(0)
	substrIndex := 0
	substrLen := len(substr)
	if substrLen == 0 {
		panic("length of substring must be bigger than zero")
	}

	var subtotal int64
	for {
		n, err := source.Read(buf)
		subtotal = 0
		for bufIndex := 0; bufIndex < n; bufIndex++ {
			if buf[bufIndex] == substr[substrIndex] {
				substrIndex++
			} else {
				substrIndex = 0
				if buf[bufIndex] == substr[substrIndex] {
					substrIndex++
				}
			}
			if substrIndex == substrLen {
				subtotal++
				substrIndex = 0
			}
		}

		total.Add(total, big.NewInt(subtotal))

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return total, err
		}
	}
	return
}

func countSubstring(source io.ReadCloser, buf, sep []byte, stop <-chan struct{}) (total *big.Int, err error) {
	defer source.Close()
	total = big.NewInt(0)
	for {
		select {
		case <-stop:
			return total, errors.New(errStopSignalStr)
		default:
			n, err := source.Read(buf)
			if n > 0 {
				count := big.NewInt(int64(bytes.Count(buf, sep)))
				total.Add(total, count)
			}
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				return total, err
			}
		}
	}
	return
}

func workerFunc(results chan<- *Result, stop <-chan struct{}, substring []byte) func(wg *sync.WaitGroup, tasks <-chan *Source) {
	return func(wg *sync.WaitGroup, tasks <-chan *Source) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in workerFunc", r)
			}
		}()
		defer wg.Done()

		buf := make([]byte, averageWebpageSize)
		for {
			select {
			case source, ok := <-tasks:
				if !ok {
					return
				}

				subtotal, err := countSubstring(source, buf, substring, stop)
				if err != nil && err.Error() == errStopSignalStr {
					return
				}

				results <- &Result{subtotal: subtotal, origin: source.origin, error: err}
			case <-stop:
				return
			}
		}
	}
}
