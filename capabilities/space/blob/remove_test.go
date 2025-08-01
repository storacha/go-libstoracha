package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripRemoveCaveats(t *testing.T) {
	digest := testutil.RandomMultihash(t)
	nb := blob.RemoveCaveats{
		Digest: digest,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.RemoveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripRemoveOk(t *testing.T) {
	size := uint64(1024)
	ok := blob.RemoveOk{
		Size: size,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := blob.RemoveOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}
