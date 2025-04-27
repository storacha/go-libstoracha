package upload_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/stretchr/testify/require"
)

func TestAddCapability(t *testing.T) {
	capability := upload.Add

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/add", capability.Can())
	})
}

func TestAddCaveatsRoundTrip(t *testing.T) {
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)

	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)

	shard2Cid, err := cid.Parse("bafybeid46f7zggioxjm5p2ze2l6s6wbqvoo4gzbdzuibgwbhe5iopu2aiy")
	require.NoError(t, err)

	t.Run("with shards", func(t *testing.T) {
		nb := upload.AddCaveats{
			Root:   rootCid,
			Shards: []cid.Cid{shard1Cid, shard2Cid},
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.AddCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb, rnb)
		require.Equal(t, nb.Root, rnb.Root)
		require.Len(t, rnb.Shards, 2)
	})

	t.Run("without shards", func(t *testing.T) {
		nb := upload.AddCaveats{
			Root: rootCid,
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.AddCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb, rnb)
		require.Equal(t, nb.Root, rnb.Root)
		require.Empty(t, rnb.Shards)
	})
}

func TestAddOkRoundTrip(t *testing.T) {
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)

	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)

	ok := upload.AddOk{
		Root:   rootCid,
		Shards: []cid.Cid{shard1Cid},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.AddOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
	require.Equal(t, ok.Root, rok.Root)
	require.Equal(t, ok.Shards, rok.Shards)
}
