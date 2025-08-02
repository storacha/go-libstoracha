package upload_test

import (
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/storacha/go-libstoracha/testutil"
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

func TestGetOkSerialization(t *testing.T) {
	ok := upload.GetOk{
		Root:       testutil.RandomCID(t),
		Shards:     []ipld.Link{testutil.RandomCID(t)},
		InsertedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt:  time.Now().UTC().Truncate(time.Second),
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := upload.GetOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Root.String(), rok.Root.String())
	require.Equal(t, len(ok.Shards), len(rok.Shards))
}
