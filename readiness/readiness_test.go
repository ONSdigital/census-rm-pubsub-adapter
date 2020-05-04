package readiness

import (
	"context"
	"os"
	"testing"
	"time"
)

var testDir = "tmpTest"
var readinessFilePath = testDir + "/test-ready"

func TestReady(t *testing.T) {
	err := setupTestDirectory()
	if err != nil {
		t.Error("Error setting up tmp test directory", err)
		return
	}

	t.Run("Test readiness file is produced and removed", testReadinessFiles)

	err = os.RemoveAll(testDir)
	if err != nil {
		t.Error(err)
	}

}

func testReadinessFiles(t *testing.T) {
	// Given
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// When
	err := Ready(ctx, readinessFilePath)
	if err != nil {
		t.Error(err)
		return
	}

	// Then
	// Check the readiness file is created
	_, err = os.Stat(readinessFilePath)
	if err != nil {
		t.Error(err)
		return
	}

	// Set a timeout for if the readiness file is not deleted
	timeoutCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	cancel()
	for {
		// Check if readiness file has been removed
		if _, err = os.Stat(readinessFilePath); err != nil {
			if os.IsNotExist(err) {
				return
			}
			t.Error(err)
			return
		}
		select {
		case <-timeoutCtx.Done():
			t.Error("Test timed out waiting for file cleanup")
			return
		default:
		}

	}
}

func setupTestDirectory() error {
	err := os.RemoveAll(testDir)
	if err != nil {
		return err
	}
	return os.Mkdir(testDir, 0700)
}
