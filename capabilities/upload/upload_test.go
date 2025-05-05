package upload_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/stretchr/testify/require"
)

func TestUploadCapability(t *testing.T) {
	capability := upload.Upload

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/*", capability.Can())
	})
}
