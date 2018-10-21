package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/borisenkova/counter/config"
	"github.com/borisenkova/counter/count"
	"github.com/borisenkova/counter/log"
)

func main() {
	log.InitLogging()
	conf := config.Load()
	tasks := make(chan *count.Source, conf.TasksBacklogSize)
	results := make(chan *count.Result, conf.ResultsBacklogSize)
	substring := conf.Substring

	httpClient := &http.Client{
		Timeout: conf.URLRequestTimeout,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	stop := make(chan struct{})
	go func() {
		<-sigChan
		close(stop)
	}()

	count.Run(os.Stdin, substring, conf.MaxNumberOfWorkers, tasks, results, stop, httpClient)
}
