package main

import (
	"net/http"
	"os"

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

	count.Run(os.Stdin, substring, conf.MaxNumberOfWorkers, tasks, results, httpClient)
}
