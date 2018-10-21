package count

import (
	"bufio"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
)

type Result struct {
	subtotal uint64
	origin   string
}

func Run(input io.Reader, substring []byte, maxNumberOfWorkers int, tasks chan *Source, results chan *Result, httpClient *http.Client) {
	workersWg := &sync.WaitGroup{}
	totalWg := &sync.WaitGroup{}
	newWorker := workerFunc(results, substring)
	workerPool := newWorkerPool(tasks, maxNumberOfWorkers, newWorker, workersWg)

	totalWg.Add(1)
	go calculateTotal(totalWg, results)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		origin := scanner.Text()
		source, err := NewSource(origin, httpClient)
		if err != nil {
			log.Printf("Can't identify Source %s: %v", origin, err)
			continue
		}

		workerPool.process(source)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error occurred on reading input: %v", err)
	}

	close(tasks)
	workersWg.Wait()
	close(results)
	totalWg.Wait()
}

func calculateTotal(wg *sync.WaitGroup, results <-chan *Result) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in calculateTotal", r)
		}
	}()
	defer wg.Done()

	total := big.NewInt(0)
	for result := range results {
		log.Printf("Count for %s: %d", result.origin, result.subtotal)
		subtotal := big.NewInt(0).SetUint64(result.subtotal)
		total.Add(total, subtotal)
	}
	log.Println("Total:", total)
}
