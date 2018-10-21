package counter

import (
	"bytes"
	"io"
	"log"
	"sync"
)

func countSubstringIn(source *Source, buf, sep []byte) (total uint64) {
	defer source.Close()
	for {
		n, err := source.Read(buf)
		if n > 0 {
			total += uint64(bytes.Count(buf, sep))
		}
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}
	}
	return
}

func workerFunc(results chan<- *Result, substring []byte) func(wg *sync.WaitGroup, tasks <-chan *Source) {
	return func(wg *sync.WaitGroup, tasks <-chan *Source) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in workerFunc", r)
			}
		}()
		defer wg.Done()

		buf := make([]byte, 2e+6)
		for {
			select {
			case source, ok := <-tasks:
				if ok {
					subtotal := countSubstringIn(source, buf, substring)
					results <- &Result{subtotal: subtotal, origin: source.origin}
				} else {
					return
				}
			}
		}
	}
}
