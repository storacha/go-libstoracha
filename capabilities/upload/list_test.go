package upload_test

import (
	"testing"
	"time"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/storacha/go-libstoracha/internal/testutil"
)

func TestListCapabilityAbility(t *testing.T) {
	capability := upload.List

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/list", capability.Can())
	})
}

func TestListCaveatsMarshaling(t *testing.T) {
	t.Run("with all parameters", func(t *testing.T) {
		cursor := "abc123"
		size := uint64(10)
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
}

func TestListOkMarshaling(t *testing.T) {
	cursor := "abc123"
	before := "before456"
	after := "after789"

	results := []upload.ListItem{
		{
			Root:       testutil.RandomCID(t),
			Shards:     []ipld.Link{testutil.RandomCID(t)},
			InsertedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt:  time.Now().UTC().Truncate(time.Second),
		},
		{
			Root:       testutil.RandomCID(t),
			Shards:     []ipld.Link{},
			InsertedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt:  time.Now().UTC().Truncate(time.Second),
		},
	}

	ok := upload.ListOk{
		Cursor:  &cursor,
		Before:  &before,
		After:   &after,
		Size:    2,
		Results: results,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.ListOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Size, rok.Size)
	require.Len(t, rok.Results, 2)
	require.Equal(t, *ok.Cursor, *rok.Cursor)
}
