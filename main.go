package main

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor/eqreceipt"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {

	ctx := context.Background()

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	a := eqreceipt.App{}

	a.Setup(ctx, "amqp://guest:guest@localhost:6672/", "project")
	go a.Consume(ctx)
	go a.Produce(ctx)

	// block until we receive a shutdown signal
	select {
	case sig := <-signals:
		log.Printf("OS Signal Received: %s", sig.String())
	}

	//cleanup for a graceful shutdown
	log.Printf("Shutting Down")

	//give the app 10 sec to cleanup before being killed
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	go func() {
		defer cancel()
		log.Printf("Rabbit Cleanup")
		a.RabbitChan.Close()
		a.RabbitConn.Close()

	}()

	//block until cancel has been called
	<-ctx.Done()

	log.Printf("Shutdown complete")
	os.Exit(1)

}
