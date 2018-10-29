package count

import (
	"context"
	"errors"
	"io"
	"log"
	"math/big"
	"sync"
)

const averageWebpageSize = 2e+6
const errCountSubstringCanceledStr = "count substring is canceled"

func countSubstring(ctx context.Context, source io.Reader, buf, substr []byte) (total *big.Int, err error) {
	total = big.NewInt(0)
	substrIndex := 0
	substrLen := len(substr)
	if substrLen == 0 {
		panic("length of substring must be bigger than zero")
	}

	var subtotal int64
	for {
		select {
		case <-ctx.Done():
			return total, errors.New(errCountSubstringCanceledStr)
		default:
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

	}
	return
}

func workerFunc(substring []byte) NewWorker {
	return func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan *Source, results chan<- *Result) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in workerFunc", r)
			}
		}()
		defer wg.Done()

		buf := make([]byte, averageWebpageSize)
		for {
			select {
			case source, hasMore := <-tasks:
				if !hasMore {
					return
				}

				subtotal, err := processSource(ctx, source, buf, substring)
				if ctx.Err() != nil {
					return
				}

				results <- &Result{subtotal: subtotal, origin: source.origin, error: err}
			case <-ctx.Done():
				return
			}
		}
	}
}

func processSource(ctx context.Context, source *Source, buf, substring []byte) (subtotal *big.Int, err error) {
	defer source.Close()
	if err = source.Load(ctx); err != nil {
		return
	}

	return countSubstring(ctx, source, buf, substring)
}
