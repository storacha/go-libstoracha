package upload_test

import (
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUploadCapability(t *testing.T) {
	capability := upload.Upload

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/*", capability.Can())
	})
}
