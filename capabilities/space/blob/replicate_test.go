package blob_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/internal/testutil"
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
		Site:     expectedLocation,
	}

	node, err := expectedNb.ToIPLD()
	require.NoError(t, err)

	actualNb, err := blob.ReplicateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedNb, actualNb)
}

func TestRoundTripReplicateOk(t *testing.T) {
	expectedTask := testutil.RandomCID(t)

	expectedOk := blob.ReplicateOk{
		Site: []blob.Promise{
			{
				UcanAwait: blob.Await{
					Selector: ".out.ok.site",
					Link:     expectedTask,
				},
			},
		},
	}

	node, err := expectedOk.ToIPLD()
	require.NoError(t, err)

	actualNb, err := blob.ReplicateOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedOk, actualNb)
}
