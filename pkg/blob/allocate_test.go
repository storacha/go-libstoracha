package blob

import (
	"testing"

	"github.com/storacha/go-capabilities/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAllocateCaveats(t *testing.T) {
	digest, bytes := testutil.RandomBytes(t, 256)
	nb := AllocateCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
		Cause: testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := AllocateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}
