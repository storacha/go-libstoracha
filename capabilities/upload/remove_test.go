package upload_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/stretchr/testify/require"
)

func TestRemoveCapability(t *testing.T) {
	capability := upload.Remove

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/remove", capability.Can())
	})
}

func TestRemoveCaveatsRoundTrip(t *testing.T) {
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	rootLink := cidlink.Link{Cid: rootCid}

	nb := upload.RemoveCaveats{
		Root: rootLink,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := upload.RemoveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Root.String(), rnb.Root.String())
}

func TestRemoveOkSerialization(t *testing.T) {
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	rootLink := cidlink.Link{Cid: rootCid}

	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	shard1Link := cidlink.Link{Cid: shard1Cid}

	ok := upload.RemoveOk{
		Root:   rootLink,
		Shards: []ipld.Link{shard1Link},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.RemoveOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Root.String(), rok.Root.String())
	require.Len(t, rok.Shards, 1)
}
