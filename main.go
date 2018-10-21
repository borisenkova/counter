package main

import (
	"net/http"
	"os"

	"github.com/borisenkova/counter/config"
	"github.com/borisenkova/counter/counter"
	"github.com/borisenkova/counter/log"
)

func main() {
	log.InitLogging()
	conf := config.Load()
	tasks := make(chan *counter.Source, conf.TasksBacklogSize)
	results := make(chan *counter.Result, conf.ResultsBacklogSize)
	substring := conf.Substring

	httpClient := &http.Client{
		Timeout: conf.URLRequestTimeout,
	}

	counter.Run(os.Stdin, substring, conf.MaxNumberOfWorkers, tasks, results, httpClient)
}
