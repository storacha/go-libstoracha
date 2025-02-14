package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAllocateCaveats(t *testing.T) {
	digest, bytes := testutil.RandomBytes(t, 256)
	nb := blob.AllocateCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: blob.Blob{
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
