package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	MaxNumberOfWorkers int
	URLRequestTimeout  time.Duration
	Substring          []byte
}

func (c *Config) loadFromEnv() *Config {
	maxNumberOfWorkersStr := os.Getenv("MAX_NUMBER_OF_WORKERS")
	if len(maxNumberOfWorkersStr) == 0 {
		maxNumberOfWorkersStr = "5"
	}
	maxNumberOfWorkers, err := strconv.Atoi(maxNumberOfWorkersStr)
	if err != nil || maxNumberOfWorkers <= 0 {
		panic(fmt.Sprintf("Invalid value for MAX_NUMBER_OF_WORKERS: must be int, got %v", maxNumberOfWorkersStr))
	}
	c.MaxNumberOfWorkers = maxNumberOfWorkers

	substring := os.Getenv("SUBSTRING")
	if len(substring) == 0 {
		substring = "Go"
	}
	c.Substring = []byte(substring)

	urlRequestTimeoutStr := os.Getenv("URL_REQUEST_TIMEOUT_MILLISECONDS")
	if len(urlRequestTimeoutStr) == 0 {
		urlRequestTimeoutStr = "60000"
	}
	urlRequestTimeout, err := strconv.ParseInt(urlRequestTimeoutStr, 10, 64)
	if err != nil || urlRequestTimeout < 1 {
		panic(fmt.Sprintf("Invalid value for URL_REQUEST_TIMEOUT_MILLISECONDS: must be uint above zero, got %v", urlRequestTimeoutStr))
	}
	c.URLRequestTimeout = time.Duration(urlRequestTimeout) * time.Millisecond

	return c
}

func Load() *Config {
	c := &Config{}
	c.loadFromEnv()
	return c
}
