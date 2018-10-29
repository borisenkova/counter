package count

import (
	"bufio"
	"context"
	"io"
	"log"
	"math/big"
	"time"
)

type Result struct {
	subtotal *big.Int
	origin   string
	error
}

func Run(ctx context.Context, input io.Reader, substring []byte, maxNumberOfWorkers int, httpTimeout time.Duration) {
	tasks := processInput(ctx, input, httpTimeout)

	pool := newWorkerPool(ctx, maxNumberOfWorkers, workerFunc(substring))
	results := pool.consume(tasks)

	calculateTotal(results)
}

func processInput(ctx context.Context, input io.Reader, httpTimeout time.Duration) <-chan *Source {
	tasks := make(chan *Source)

	go func() {
		scanner := bufio.NewScanner(input)
		defer func() {
			if err := scanner.Err(); err != nil {
				log.Printf("Error occurred on reading input: %v", err)
			}
		}()
		defer close(tasks)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if hasMore := scanner.Scan(); !hasMore {
					return
				}

				origin := scanner.Text()
				source, err := NewSource(origin, httpTimeout)
				if err != nil {
					log.Printf("Can't identify source '%s': %v", origin, err)
					continue
				}

				tasks <- source
			}
		}
	}()

	return tasks
}

func calculateTotal(results <-chan *Result) {
	total := big.NewInt(0)
	for result := range results {
		if result.error != nil {
			log.Printf("Count for %s: %v, and got error while processing: %v", result.origin, result.subtotal, result.error)
			continue
		}

		log.Printf("Count for %s: %v", result.origin, result.subtotal)
		total.Add(total, result.subtotal)
	}

	log.Println("Total:", total)
}
