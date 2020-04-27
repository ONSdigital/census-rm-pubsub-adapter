package main

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	appConfig, err := config.GetConfig()
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error getting config at startup"))
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Trap SIGINT to trigger eqReceiptProcessor graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Start EQ receipt processing
	eqReceiptProcessor := processor.NewEqReceiptProcessor(ctx, appConfig)
	go eqReceiptProcessor.Consume(ctx)
	go eqReceiptProcessor.Process(ctx)

	// Start offline receipt processing
	offlineReceiptProcessor := processor.NewOfflineReceiptProcessor(ctx, appConfig)
	go offlineReceiptProcessor.Consume(ctx)
	go offlineReceiptProcessor.Process(ctx)

	// Start PPO undelivered processing
	ppoUndeliveredProcessor := processor.NewPpoUndeliveredProcessor(ctx, appConfig)
	go ppoUndeliveredProcessor.Consume(ctx)
	go ppoUndeliveredProcessor.Process(ctx)

	// Start QM undelivered processing
	qmUndeliveredProcessor := processor.NewQmUndeliveredProcessor(ctx, appConfig)
	go qmUndeliveredProcessor.Consume(ctx)
	go qmUndeliveredProcessor.Process(ctx)

	// block until we receive eqReceiptProcessor shutdown signal
	select {
	case sig := <-signals:
		log.Printf("OS Signal Received: %s", sig.String())
	}

	//cleanup for eqReceiptProcessor graceful shutdown
	log.Printf("Shutting Down")

	//give the app 10 sec to cleanup before being killed
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)

	go func() {
		//send cancel to all consumers
		cancel()
		//defer shutdownCancel - this will be called once all things are close or when the timeout is reached
		defer shutdownCancel()
		log.Printf("Rabbit Cleanup")
		err = eqReceiptProcessor.RabbitChan.Close()
		if err != nil {
			log.Println(errors.Wrap(err, "Error closing rabbit channel during shutdown"))
		}
		err = eqReceiptProcessor.RabbitConn.Close()
		if err != nil {
			log.Println(errors.Wrap(err, "Error closing rabbit connection during shutdown"))
		}

	}()

	//block until shutdown cancel has been called
	<-shutdownCtx.Done()

	log.Printf("Shutdown complete")
	os.Exit(1)

}
