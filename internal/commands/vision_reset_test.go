package commands

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestVisionResetCommand(t *testing.T) {
	t.Run("ResetCaptionAndLabels", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("VisionResetTarget")

		args := []string{
			"reset",
			"--models=caption,labels",
			"--source=ollama",
			"--yes",
			fmt.Sprintf("uid:%s", fixture.PhotoUID),
		}

		if output, err := RunWithTestContext(VisionResetCommand, args); err != nil {
			t.Fatalf("%T: %v", err, err)
		} else {
			assert.Empty(t, output)
		}
	})
}
