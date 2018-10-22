package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	MaxNumberOfWorkers int
	TasksBacklogSize   int
	ResultsBacklogSize int
	URLRequestTimeout  time.Duration
	Substring          []byte
}

func (c *Config) loadFromEnv() *Config {
	maxNumberOfWorkersStr := os.Getenv("MAX_NUMBER_OF_WORKERS")
	if len(maxNumberOfWorkersStr) == 0 {
		maxNumberOfWorkersStr = "5"
	}
	maxNumberOfWorkers, err := strconv.Atoi(maxNumberOfWorkersStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid value for MAX_NUMBER_OF_WORKERS: must be int, got %v", maxNumberOfWorkersStr))
	}
	c.MaxNumberOfWorkers = maxNumberOfWorkers

	tasksBacklogSizeStr := os.Getenv("TASKS_BACKLOG_SIZE")
	if len(tasksBacklogSizeStr) == 0 {
		tasksBacklogSizeStr = "3"
	}
	tasksBacklogSize, err := strconv.Atoi(tasksBacklogSizeStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid value for TASKS_BACKLOG_SIZE: must be int, got %v", tasksBacklogSizeStr))
	}
	c.TasksBacklogSize = tasksBacklogSize

	resultsBacklogSizeStr := os.Getenv("RESULTS_BACKLOG_SIZE")
	if len(resultsBacklogSizeStr) == 0 {
		resultsBacklogSizeStr = "3"
	}
	resultsBacklogSize, err := strconv.Atoi(resultsBacklogSizeStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid value for RESULTS_BACKLOG_SIZE: must be int, got %v", resultsBacklogSizeStr))
	}
	c.ResultsBacklogSize = resultsBacklogSize

	substring := os.Getenv("SUBSTRING")
	if len(substring) == 0 {
		substring = "Go"
	}
	c.Substring = []byte(substring)

	urlRequestTimeoutStr := os.Getenv("URL_REQUEST_TIMEOUT_SECONDS")
	if len(urlRequestTimeoutStr) == 0 {
		urlRequestTimeoutStr = "10"
	}
	urlRequestTimeout, err := strconv.ParseInt(urlRequestTimeoutStr, 10, 64)
	if err != nil || urlRequestTimeout < 1 {
		panic(fmt.Sprintf("Invalid value for URL_REQUEST_TIMEOUT_SECONDS: must be uint above zero, got %v", urlRequestTimeoutStr))
	}
	c.URLRequestTimeout = time.Duration(urlRequestTimeout) * time.Second

	return c
}

func Load() *Config {
	c := &Config{}
	c.loadFromEnv()
	return c
}
