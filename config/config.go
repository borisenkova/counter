package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	MaxNumberOfWorkers    int
	URLRequestTimeout     time.Duration
	WorkerShutdownTimeout time.Duration
	Substring             []byte
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

	c.URLRequestTimeout = parseDurationMs("URL_REQUEST_TIMEOUT_MILLISECONDS", "60000")
	c.WorkerShutdownTimeout = parseDurationMs("WORKER_SHUTDOWN_TIMEOUT_MILLISECONDS", "60000")

	return c
}

func parseDurationMs(envVar string, defaultValue string) time.Duration {
	durationStr := os.Getenv(envVar)
	if len(durationStr) == 0 {
		durationStr = defaultValue
	}

	duration, err := strconv.ParseInt(durationStr, 10, 64)
	if err != nil || duration < 1 {
		panic(fmt.Sprintf("Invalid value for %s: must be uint above zero, got %v", envVar, durationStr))
	}

	return time.Duration(duration) * time.Millisecond
}

func Load() *Config {
	c := &Config{}
	c.loadFromEnv()
	return c
}
