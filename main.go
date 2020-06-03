package main

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/logger"
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

	// Configure the logger
	err = logger.ConfigureLogger(appConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to configure logger at startup"))
	}
	logger.Logger.Infow("Launching PubSub Adapter")

	// Channel for goroutines to notify main of errors
	errChan := make(chan error)

	processors, err := StartProcessors(ctx, appConfig, errChan)
	if err != nil {
		logger.Logger.Errorw("Error starting processors", "error", err)
		shutdown(ctx, cancel, processors)
	}

	RunUntilUnrecoverableErrorOrShutdownSignal(ctx, appConfig, cancel, errChan, processors)
}

func RunUntilUnrecoverableErrorOrShutdownSignal(ctx context.Context, appConfig *config.Configuration, cancel context.CancelFunc, errChan chan error, processors []*processor.Processor) {
	// Trap SIGINT to trigger graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	////Indicate readiness
	//err := readiness.Ready(ctx, appConfig.ReadinessFilePath)
	//if err != nil {
	//	logger.Logger.Errorw("Error indicating readiness", "error", err)
	//	shutdown(ctx, cancel, processors)
	//}

	shuttingDown := false

	// Block until we receive OS shutdown signal or error
	// shutdown if OS Signal, or if processors won't restart
	select {
	case sig := <-signals:
		logger.Logger.Infow("OS Signal Received", "signal", sig.String())
		shuttingDown = true
		shutdown(ctx, cancel, processors)

	case err := <-errChan:
		logger.Logger.Errorw("Error Received", "error", err)

		//Only try restarting the Processors if we're not shutting down, else we could loop.
		if !shuttingDown {
			TryRestartBrokenProcessors(processors, ctx, appConfig, errChan)
		}

	}
}

func TryRestartBrokenProcessors(processors []*processor.Processor, ctx context.Context, appConfig *config.Configuration, errChan chan error) {
	// Todo find the correct one to restart, and check if that one doesn't err?
	// Some sort of time out? Spawn new Context witn Timeout (try google) or try 'retry library'
	// Processor to have Restart function, which shuts down, then restarts it all
	// Rather than array of processor have a map, and return 'key' in errChan msg

	for _, p := range processors {
		if p.Errored {
			//This is the one to restart, and we can get the details?
			//This is sort of a nice to have
			p.ShutdownProcessor(ctx)

			_, err := p.StartProcessor(ctx, appConfig)

			if err != nil {
				logger.Logger.Errorw("Error Received", "error", err)
				//return nil, errors.Wrap(err, "Error Trying to restart Processor")
			}
		}
	}
}

func StartProcessors(ctx context.Context, cfg *config.Configuration, errChan chan error) ([]*processor.Processor, error) {
	processors := make([]*processor.Processor, 0)

	// Start EQ receipt processing
	eqReceiptProcessor, err := processor.NewEqReceiptProcessor(ctx, cfg, errChan)
	if err != nil {
		return nil, errors.Wrap(err, "Error starting eQ receipt processor")
	}
	processors = append(processors, eqReceiptProcessor)

	// Start offline receipt processing
	offlineReceiptProcessor, err := processor.NewOfflineReceiptProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting offline receipt processor")
	}
	processors = append(processors, offlineReceiptProcessor)

	// Start PPO undelivered processing
	ppoUndeliveredProcessor, err := processor.NewPpoUndeliveredProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting PPO undelivered processor")
	}
	processors = append(processors, ppoUndeliveredProcessor)

	// Start QM undelivered processing
	qmUndeliveredProcessor, err := processor.NewQmUndeliveredProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting QM undelivered processor")
	}
	processors = append(processors, qmUndeliveredProcessor)

	return processors, nil
}

func shutdown(ctx context.Context, cancel context.CancelFunc, processors []*processor.Processor) {
	// cleanup for graceful shutdown
	logger.Logger.Info("Shutting Down")

	// give the app 10 sec to cleanup before being killed
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)

	go func() {
		// send cancel to all consumers
		cancel()
		// this will be called once cleanup completes or when the timeout is reached
		defer shutdownCancel()

		logger.Logger.Info("Starting rabbit cleanup")
		for _, p := range processors {
			p.CloseRabbit()
		}

	}()

	//block until shutdown cancel has been called
	<-shutdownCtx.Done()

	logger.Logger.Info("Shutdown complete")
	os.Exit(1)
}
