package readiness

import (
	"context"
	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)
	}

}

func testReadinessFiles(t *testing.T) {
	// Given
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// When
	err := Ready(ctx, readinessFilePath)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Then
	// Check the readiness file is created
	_, err = os.Stat(readinessFilePath)
	if !assert.NoError(t, err, "Readiness file was not present") {
		return
	}

	// Set a timeout for if the readiness file is not deleted
	timeoutCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	// Trigger the cancel which should result in the file being removed
	cancel()

	for {
		// Check if readiness file has been removed
		if _, err = os.Stat(readinessFilePath); err != nil {
			// Succeed if the file does not now exist
			if os.IsNotExist(err) {
				return
			}
			// Error the test on any other error
			assert.NoError(t, err)
			return
		}

		// Fail the test if it times out before the file is removed
		select {
		case <-timeoutCtx.Done():
			assert.Fail(t, "Test timed out waiting for file cleanup")
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
