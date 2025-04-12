package blob_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

func TestRoundTripReplicateCaveats(t *testing.T) {
	expectedSize := uint64(256)
	expectedReplicas := uint(8)
	expectedLocation := testutil.RandomCID(t)
	expectedDigest, _ := testutil.RandomBytes(t, int(expectedSize))

	expectedNb := blob.ReplicateCaveats{
		Blob: types.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Replicas: expectedReplicas,
		Location: expectedLocation,
	}

	node, err := expectedNb.ToIPLD()
	require.NoError(t, err)

	actualNb, err := blob.ReplicateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedNb, actualNb)
}
