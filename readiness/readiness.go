package readiness

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
)

func Ready(ctx context.Context, readinessFilePath string) error {
	_, err := os.Stat(readinessFilePath)
	if err == nil {
		// TODO Log a warning that the readiness file already existed
		log.Println("Readiness file already existed")
	}
	_, err = os.Create(readinessFilePath)
	if err != nil {
		return err
	}
	go removeReadyWhenDone(ctx, readinessFilePath)

	return nil
}

func removeReadyWhenDone(ctx context.Context, readinessFilePath string) {
	select {
	case <-ctx.Done():
		log.Println("Removing readiness file")
		err := os.Remove(readinessFilePath)
		if err != nil {
			log.Println(errors.Wrap(err, fmt.Sprintf("Error removing readiness file: %s", readinessFilePath)))
		}
	}
}
