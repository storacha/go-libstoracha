package upload_test

import (
	"testing"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/upload"
)

func TestRemoveCapability(t *testing.T) {
	capability := upload.Remove

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/remove", capability.Can())
	})
}

func TestRemoveCaveatsRoundTrip(t *testing.T) {
	nb := upload.RemoveCaveats{
		Root: testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := upload.RemoveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Root.String(), rnb.Root.String())
}

func TestRemoveOkSerialization(t *testing.T) {
	ok := upload.RemoveOk{
		Root:   testutil.RandomCID(t),
		Shards: []ipld.Link{testutil.RandomCID(t)},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.RemoveOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Root.String(), rok.Root.String())
	require.Len(t, rok.Shards, 1)
}
