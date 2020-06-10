package main

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/config"
	"github.com/ONSdigital/census-rm-pubsub-adapter/logger"
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

	// Configure the logger
	err = logger.ConfigureLogger(appConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to configure logger at startup"))
	}
	logger.Logger.Infow("Launching PubSub Adapter")

	// Trap SIGINT to trigger graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Channel for goroutines to notify main of errors
	errChan := make(chan processor.Error)

	processors, err := StartProcessors(ctx, appConfig, errChan)

	// Ensure proper shutdown is attempted
	defer shutdown(ctx, cancel, processors)

	if err != nil {
		logger.Logger.Errorw("Error starting processors", "error", err)
		return
	}

	// Wait for processors to start properly
	if err := waitForStartup(ctx, errChan); err != nil {
		logger.Logger.Info("Shutting down due to start up error")
		return
	}
	// Indicate readiness
	err = readiness.Ready(ctx, appConfig.ReadinessFilePath)
	if err != nil {
		logger.Logger.Errorw("Error indicating readiness", "error", err)
		return
	}

	// Block until we receive OS shutdown signal or error
	RunLoop(ctx, appConfig, signals, errChan)
}

func RunLoop(ctx context.Context, cfg *config.Configuration, signals chan os.Signal, errChan chan processor.Error) {
	restartTimeout := context.Background()
	allRunning := true
	for {
		select {
		case sig := <-signals:
			logger.Logger.Infow("OS Signal Received", "signal", sig.String())
			return
		case <-restartTimeout.Done():
			logger.Logger.Infow("Could not successfully restart, shutting down", "restartTimeout", cfg.RestartTimeout)
			return
		case processorErr := <-errChan:
			logger.Logger.Errorw("Processor error received", "error", processorErr.Err, "processor", processorErr.Name)

			// If the app was running, start a timeout
			if allRunning {
				restartTimeout, _ = context.WithTimeout(context.Background(), time.Duration(cfg.RestartTimeout)*time.Second)
				allRunning = false
			}

			//if err := readiness.Unready(cfg.ReadinessFilePath); err != nil {
			//	logger.Logger.Errorw("Error showing not ready", "error", err)
			//}

			processorErr.Restart(ctx)

			if err := waitForStartup(ctx, errChan); err != nil {
				go processorErr.QueueError(err)
			} else {
				allRunning = true
				restartTimeout = context.Background()
			}

			// Limit the rate of restarts
			time.Sleep(1 * time.Second)

			//if err := readiness.Ready(ctx, cfg.ReadinessFilePath); err != nil {
			//	logger.Logger.Errorw("Error showing ready", "error", err)
			//}
		}
	}
}

func StartProcessors(ctx context.Context, cfg *config.Configuration, errChan chan processor.Error) ([]*processor.Processor, error) {
	processors := make([]*processor.Processor, 0)

	// Initialise EQ receipt processing
	eqReceiptProcessor, err := processor.NewEqReceiptProcessor(ctx, cfg, errChan)
	if err != nil {
		return nil, errors.Wrap(err, "Error starting eQ receipt processor")
	}
	processors = append(processors, eqReceiptProcessor)

	// Initialise offline receipt processing
	offlineReceiptProcessor, err := processor.NewOfflineReceiptProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting offline receipt processor")
	}
	processors = append(processors, offlineReceiptProcessor)

	// Initialise PPO undelivered processing
	ppoUndeliveredProcessor, err := processor.NewPpoUndeliveredProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting PPO undelivered processor")
	}
	processors = append(processors, ppoUndeliveredProcessor)

	// Initialise QM undelivered processing
	qmUndeliveredProcessor, err := processor.NewQmUndeliveredProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting QM undelivered processor")
	}
	processors = append(processors, qmUndeliveredProcessor)

	// Initialise fulfilment confirmed processing
	fulfilmentConfirmedProcessor, err := processor.NewFulfilmentConfirmedProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting fulfilment confirmed processor")
	}
	processors = append(processors, fulfilmentConfirmedProcessor)

	eqFulfilmentProcessor, err := processor.NewEqFulfilmentProcessor(ctx, cfg, errChan)
	if err != nil {
		return processors, errors.Wrap(err, "Error starting eq fulfilment processor")
	}
	processors = append(processors, eqFulfilmentProcessor)

	return processors, nil
}

func waitForStartup(ctx context.Context, errChan chan processor.Error) error {
	startupTimer, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	select {
	case processorError := <-errChan:
		processorError.Logger.Errorw("Processor errored during startup period", "error", processorError.Err)
		return processorError.Err
	case <-startupTimer.Done():
		logger.Logger.Debug("Startup complete")
		return nil
	}
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
			p.CloseRabbit(false)
		}

	}()

	//block until shutdown cancel has been called
	<-shutdownCtx.Done()

	logger.Logger.Info("Shutdown complete")
	os.Exit(1)
}
