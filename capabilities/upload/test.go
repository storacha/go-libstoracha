package upload_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAddCaveats(t *testing.T) {
	root := testutil.RandomCID(t)
	shard1 := testutil.RandomCID(t)
	shard2 := testutil.RandomCID(t)
	
	nb := upload.AddCaveats{
		Root:   root,
		Shards: []cid.Cid{shard1, shard2},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := upload.AddCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripAddOk(t *testing.T) {
	root := testutil.RandomCID(t)
	shard1 := testutil.RandomCID(t)
	shard2 := testutil.RandomCID(t)
	
	ok := upload.AddOk{
		Root:   root,
		Shards: []cid.Cid{shard1, shard2},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.AddOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestRoundTripGetCaveats(t *testing.T) {
	root := testutil.RandomCID(t)
	
	nb := upload.GetCaveats{
		Root:   &root,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := upload.GetCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripGetOk(t *testing.T) {
	root := testutil.RandomCID(t)
	shard1 := testutil.RandomCID(t)
	
	ok := upload.GetOk{
		Root:       root,
		Shards:     []cid.Cid{shard1},
		InsertedAt: "2023-01-01T00:00:00Z",
		UpdatedAt:  "2023-01-02T00:00:00Z",
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.GetOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestRoundTripRemoveCaveats(t *testing.T) {
	root := testutil.RandomCID(t)
	
	nb := upload.RemoveCaveats{
		Root: root,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := upload.RemoveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripRemoveOk(t *testing.T) {
	root := testutil.RandomCID(t)
	shard1 := testutil.RandomCID(t)
	
	ok := upload.RemoveOk{
		Root:   root,
		Shards: []cid.Cid{shard1},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.RemoveOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestRoundTripListCaveats(t *testing.T) {
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
	require.Equal(t, nb, rnb)
}

func TestRoundTripListOk(t *testing.T) {
	cursor := "abc123"
	before := "before456"
	after := "after789"
	root1 := testutil.RandomCID(t)
	root2 := testutil.RandomCID(t)
	shard1 := testutil.RandomCID(t)
	
	ok := upload.ListOk{
		Cursor: &cursor,
		Before: &before,
		After:  &after,
		Size:   2,
		Results: []upload.ListItem{
			{
				Root:       root1,
				Shards:     []cid.Cid{shard1},
				InsertedAt: "2023-01-01T00:00:00Z",
				UpdatedAt:  "2023-01-02T00:00:00Z",
			},
			{
				Root:       root2,
				InsertedAt: "2023-01-03T00:00:00Z",
				UpdatedAt:  "2023-01-04T00:00:00Z",
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.ListOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestCapabilityDerivation(t *testing.T) {

	
	t.Run("UploadAddFromUploadStar", func(t *testing.T) {

	})
	
	t.Run("UploadGetWithRootConstraint", func(t *testing.T) {
	})
	
}