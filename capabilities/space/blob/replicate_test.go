package blob_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	blob2 "github.com/storacha/go-libstoracha/capabilities/space/blob"
)

func TestRoundTripReplicateCaveats(t *testing.T) {
	expectedSize := uint64(256)
	expectedReplicas := uint(8)
	expectedLocation := testutil.RandomCID(t)
	expectedDigest, _ := testutil.RandomBytes(t, int(expectedSize))

	expectedNb := blob2.ReplicateCaveats{
		Blob: blob.Blob{
			Digest: expectedDigest,
			Size:   expectedSize,
		},
		Replicas: expectedReplicas,
		Location: expectedLocation,
	}

	node, err := expectedNb.ToIPLD()
	require.NoError(t, err)

	actualNb, err := blob2.ReplicateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedNb, actualNb)
}
