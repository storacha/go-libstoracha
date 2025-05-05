package upload_test

import (
	"testing"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/upload"
)

func TestAddCapability(t *testing.T) {
	capability := upload.Add

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/add", capability.Can())
	})
}

func TestAddCaveatsRoundTrip(t *testing.T) {
	t.Run("with shards", func(t *testing.T) {
		nb := upload.AddCaveats{
			Root:   testutil.RandomCID(t),
			Shards: []ipld.Link{testutil.RandomCID(t), testutil.RandomCID(t)},
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.AddCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb.Root.String(), rnb.Root.String())
		require.Len(t, rnb.Shards, 2)
	})

	t.Run("without shards", func(t *testing.T) {
		nb := upload.AddCaveats{
			Root:   testutil.RandomCID(t),
			Shards: []ipld.Link{},
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.AddCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb.Root.String(), rnb.Root.String())
		require.Empty(t, rnb.Shards)
	})
}

func TestAddOkSerialization(t *testing.T) {
	ok := upload.AddOk{
		Root:   testutil.RandomCID(t),
		Shards: []ipld.Link{testutil.RandomCID(t)},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.AddOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Root.String(), rok.Root.String())
	require.Equal(t, len(ok.Shards), len(rok.Shards))
}
