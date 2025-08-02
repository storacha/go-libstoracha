package blobindex_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/blobindex"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestFromToArchive(t *testing.T) {
	root, _, contentCarBytes := testutil.RandomCAR(t, 32)
	index, err := blobindex.FromShardArchives(root, [][]byte{contentCarBytes})
	require.NoError(t, err)
	r, err := index.Archive()
	require.NoError(t, err)
	newIndex, err := blobindex.Extract(r)
	require.NoError(t, err)
	require.Equal(t, root.String(), newIndex.Content().String())
	require.NotZero(t, newIndex.Shards().Size())
	require.Equal(t, index.Shards().Size(), newIndex.Shards().Size())
	for key, shard := range newIndex.Shards().Iterator() {
		require.True(t, index.Shards().Has(key))
		require.Equal(t, index.Shards().Get(key).Size(), shard.Size())
	}
}
