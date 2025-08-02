package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/testutil"

	"github.com/stretchr/testify/require"
)

func TestRoundTripAllocateCaveats(t *testing.T) {
	bytes := testutil.RandomBytes(t, 256)
	digest := testutil.MultihashFromBytes(t, bytes)
	nb := blob.AllocateCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: types.Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
		Cause: testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.AllocateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}
