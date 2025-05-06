package replica_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/blob/replica"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/internal/testutil"
)

func TestRoundTripAllocateCaveats(t *testing.T) {
	expectedSpace := testutil.RandomPrincipal(t).DID()
	expectedSize := 256
	expectedDigest, _ := testutil.RandomBytes(t, expectedSize)
	expectedLocation := testutil.RandomCID(t)
	expectedCause := testutil.RandomCID(t)

	expectedNp := replica.AllocateCaveats{
		Space: expectedSpace,
		Blob: types.Blob{
			Digest: expectedDigest,
			Size:   uint64(expectedSize),
		},
		Location: expectedLocation,
		Cause:    expectedCause,
	}

	node, err := expectedNp.ToIPLD()
	require.NoError(t, err)

	actualNp, err := replica.AllocateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedNp, actualNp)

}
