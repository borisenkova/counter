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
	subtotal *big.Int
	origin   string
	error
}

func Run(input io.Reader, substring []byte, maxNumberOfWorkers int, tasks chan *Source, results chan *Result, stop <-chan struct{}, httpClient *http.Client) {
	workersWg := &sync.WaitGroup{}
	totalWg := &sync.WaitGroup{}
	stopInput := make(chan struct{})
	inputWg := &sync.WaitGroup{}
	stopWorkers := make(chan struct{})

	go func() {
		<-stop
		close(stopInput)
		inputWg.Wait()
		close(stopWorkers)
	}()

	newWorker := workerFunc(results, stopWorkers, substring)
	workerPool := newWorkerPool(tasks, maxNumberOfWorkers, newWorker, workersWg)

	totalWg.Add(1)
	go calculateTotal(totalWg, results)

	scanner := bufio.NewScanner(input)

	inputWg.Add(1)
	stopped := processInput(scanner, workerPool, httpClient, stopInput, inputWg)
	if err := scanner.Err(); err != nil {
		log.Printf("Error occurred on reading input: %v", err)
	}

	if !stopped {
		close(tasks)
	}

	workersWg.Wait()
	close(results)
	totalWg.Wait()
}

func processInput(scanner *bufio.Scanner, pool *WorkerPool, client *http.Client, stop <-chan struct{}, wg *sync.WaitGroup) (stopped bool) {
	defer wg.Done()
	for {
		select {
		case <-stop:
			return true
		default:
			if hasMore := scanner.Scan(); !hasMore {
				return false
			}
			origin := scanner.Text()
			source, err := NewSource(origin, client)
			if err != nil {
				log.Printf("Can't identify source '%s': %v", origin, err)
				continue
			}
			pool.process(source, stop)
		}
	}
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
		if result.error != nil {
			log.Printf("Count for %s: %v, and got error while processing: %v", result.origin, result.subtotal, result.error)
			continue
		}

		log.Printf("Count for %s: %v", result.origin, result.subtotal)
		total.Add(total, result.subtotal)
	}
	log.Println("Total:", total)
}
