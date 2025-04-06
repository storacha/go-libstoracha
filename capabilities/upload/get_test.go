package upload_test

import (
	"testing"

	"github.com/ipfs/go-cid"
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
		
		nb := upload.GetCaveats{
			Root: &rootCid,
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.GetCaveatsReader.Read(node)
		require.NoError(t, err)
		require.NotNil(t, rnb.Root)
	})

}

func TestGetOkRoundTrip(t *testing.T) {
	rootCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	
	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	
	ok := upload.GetOk{
		Root:       rootCid,
		Shards:     []cid.Cid{shard1Cid},
		InsertedAt: "2023-01-01T00:00:00Z",
		UpdatedAt:  "2023-01-02T00:00:00Z",
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.GetOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.InsertedAt, rok.InsertedAt)
	require.Equal(t, ok.UpdatedAt, rok.UpdatedAt)
}