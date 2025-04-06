package upload_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/stretchr/testify/require"
)

func TestListCapability(t *testing.T) {
	capability := upload.List

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/list", capability.Can())
	})
}

func TestListCaveatsRoundTrip(t *testing.T) {
	t.Run("with all parameters", func(t *testing.T) {
		cursor := "abc123"
		size := 10
		pre := true
		
		nb := upload.ListCaveats{
			Cursor: &cursor,
			Size:   &size,
			Pre:    &pre,
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.ListCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, cursor, *rnb.Cursor)
		require.Equal(t, size, *rnb.Size)
		require.Equal(t, pre, *rnb.Pre)
	})

	t.Run("without parameters", func(t *testing.T) {
		nb := upload.ListCaveats{}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := upload.ListCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Nil(t, rnb.Cursor)
		require.Nil(t, rnb.Size)
		require.Nil(t, rnb.Pre)
	})
}

func TestListOkRoundTrip(t *testing.T) {
	cursor := "abc123"
	before := "before456"
	after := "after789"
	
	rootCid1, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	
	rootCid2, err := cid.Parse("bafybeies3cfa2dlg6pfkuoo7lbdkphpsgpjj7ivyfxs6han37qawtx5inq")
	require.NoError(t, err)
	
	shard1Cid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	
	ok := upload.ListOk{
		Cursor: &cursor,
		Before: &before,
		After:  &after,
		Size:   2,
		Results: []upload.ListItem{
			{
				Root:       rootCid1,
				Shards:     []cid.Cid{shard1Cid},
				InsertedAt: "2023-01-01T00:00:00Z",
				UpdatedAt:  "2023-01-02T00:00:00Z",
			},
			{
				Root:       rootCid2,
				InsertedAt: "2023-01-03T00:00:00Z",
				UpdatedAt:  "2023-01-04T00:00:00Z",
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.ListOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Size, rok.Size)
	require.Len(t, rok.Results, 2)
	require.Equal(t, *ok.Cursor, *rok.Cursor)
}