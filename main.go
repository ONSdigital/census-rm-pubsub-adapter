package main

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/processor"
	"github.com/ONSdigital/census-rm-pubsub-adapter/readiness"
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

	processors, err := StartProcessors(ctx, appConfig)
	if err != nil {
		shutdown(ctx, cancel, processors)
	}
	// Indicate readiness
	err = readiness.Ready(ctx, appConfig.ReadinessFilePath)
	if err != nil {
		shutdown(ctx, cancel, processors)
	}
	log.Println("Pubsub Adapter successfully started")

	// block until we receive eqReceiptProcessor shutdown signal
	select {
	case sig := <-signals:
		log.Printf("OS Signal Received: %s", sig.String())
	}

	shutdown(ctx, cancel, processors)

}

func StartProcessors(ctx context.Context, cfg *config.Configuration) ([]*processor.Processor, error) {
	processors := make([]*processor.Processor, 0)

	// Start EQ receipt processing
	eqReceiptProcessor, err := processor.NewEqReceiptProcessor(ctx, cfg)
	if err != nil {
		return nil, err
	}
	processors = append(processors, eqReceiptProcessor)

	// Start offline receipt processing
	offlineReceiptProcessor, err := processor.NewOfflineReceiptProcessor(ctx, cfg)
	if err != nil {
		return nil, err
	}
	processors = append(processors, offlineReceiptProcessor)

	// Start PPO undelivered processing
	ppoUndeliveredProcessor, err := processor.NewPpoUndeliveredProcessor(ctx, cfg)
	if err != nil {
		return nil, err
	}
	processors = append(processors, ppoUndeliveredProcessor)

	// Start QM undelivered processing
	qmUndeliveredProcessor, err := processor.NewQmUndeliveredProcessor(ctx, cfg)
	if err != nil {
		return nil, err
	}
	processors = append(processors, qmUndeliveredProcessor)
	return processors, nil
}

func shutdown(ctx context.Context, cancel context.CancelFunc, processors []*processor.Processor) {
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
		for _, p := range processors {
			p.CloseRabbit()
		}

	}()

	//block until shutdown cancel has been called
	<-shutdownCtx.Done()

	log.Printf("CloseRabbit complete")
	os.Exit(1)
}
