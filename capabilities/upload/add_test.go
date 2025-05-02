package upload_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/storacha/go-ucanto/core/ipld"
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
	rootLink := cidlink.Link{Cid: rootCid}

	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	shard1Link := cidlink.Link{Cid: shard1Cid}

	shard2Cid, err := cid.Parse("bafybeid46f7zggioxjm5p2ze2l6s6wbqvoo4gzbdzuibgwbhe5iopu2aiy")
	require.NoError(t, err)
	shard2Link := cidlink.Link{Cid: shard2Cid}

	t.Run("with shards", func(t *testing.T) {
		nb := upload.AddCaveats{
			Root:   rootLink,
			Shards: []ipld.Link{shard1Link, shard2Link},
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
			Root:   rootLink,
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
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	rootLink := cidlink.Link{Cid: rootCid}

	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	shard1Link := cidlink.Link{Cid: shard1Cid}

	ok := upload.AddOk{
		Root:   rootLink,
		Shards: []ipld.Link{shard1Link},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.AddOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Root.String(), rok.Root.String())
	require.Equal(t, len(ok.Shards), len(rok.Shards))
}
