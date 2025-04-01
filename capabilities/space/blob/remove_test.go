package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripRemoveCaveats(t *testing.T) {
	digest, _ := testutil.RandomBytes(t, 256)
	nb := blob.RemoveCaveats{
		Digest: digest,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.RemoveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}
