package main

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	a := processor.New(ctx, "amqp://guest:guest@localhost:6672/", "project")
	go a.Consume(ctx)
	go a.Process(ctx)

	// block until we receive a shutdown signal
	select {
	case sig := <-signals:
		log.Printf("OS Signal Received: %s", sig.String())
	}

	//cleanup for a graceful shutdown
	log.Printf("Shutting Down")

	//give the app 10 sec to cleanup before being killed
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)

	go func() {
		//send cancel to all consumers
		cancel()
		//defer shutdownCancel - this will be called once all things are close or when the timeout is reached
		defer shutdownCancel()
		log.Printf("Rabbit Cleanup")
		a.RabbitChan.Close()
		a.RabbitConn.Close()

	}()

	//block until cancel has been called
	<-shutdownCtx.Done()

	log.Printf("Shutdown complete")
	os.Exit(1)

}
