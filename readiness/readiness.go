package readiness

import (
	"context"
	"github.com/ONSdigital/census-rm-pubsub-adapter/logger"
	"os"
)

func Ready(ctx context.Context, readinessFilePath string) error {
	_, err := os.Stat(readinessFilePath)
	if err == nil {
		logger.Logger.Errorw("Readiness file already existed", "readinessFilePath", readinessFilePath)
	}
	_, err = os.Create(readinessFilePath)
	if err != nil {
		return err
	}
	go removeReadyWhenDone(ctx, readinessFilePath)

	return nil
}

func Unready(readinessFilePath string) error {
	return os.Remove(readinessFilePath)
}

func removeReadyWhenDone(ctx context.Context, readinessFilePath string) {
	<-ctx.Done()
	logger.Logger.Info("Removing readiness file")
	err := Unready(readinessFilePath)
	if err != nil {
		logger.Logger.Errorw("Error removing readiness file", "readinessFilePath", readinessFilePath, "error", err)
	}
}
