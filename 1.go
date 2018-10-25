package main

import (
	"context"
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
	substring := conf.Substring

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigChan
		cancel()
	}()

	count.Run(ctx, os.Stdin, substring, conf.MaxNumberOfWorkers, conf.URLRequestTimeout)
}
