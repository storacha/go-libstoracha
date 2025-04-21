package upload_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/stretchr/testify/require"
)

func TestGetCapability(t *testing.T) {
	capability := upload.Get

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/get", capability.Can())
	})
}

func TestGetCaveatsRoundTrip(t *testing.T) {
	t.Run("with root", func(t *testing.T) {
		rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
		require.NoError(t, err)
		rootLink := cidlink.Link{Cid: rootCid}

		nb := upload.GetCaveats{
			Root: rootLink,
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.GetCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb.Root.String(), rnb.Root.String())
	})
}

func TestGetOkRoundTrip(t *testing.T) {
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	rootLink := cidlink.Link{Cid: rootCid}

	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	shard1Link := cidlink.Link{Cid: shard1Cid}

	ok := upload.GetOk{
		Root:   rootLink,
		Shards: []datamodel.Link{shard1Link},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.GetOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Root.String(), rok.Root.String())
	require.Equal(t, len(ok.Shards), len(rok.Shards))
}
