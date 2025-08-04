package replica_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/blob/replica"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/testutil"
)

func TestRoundTripAllocateCaveats(t *testing.T) {
	expectedSpace := testutil.RandomPrincipal(t).DID()
	expectedSize := 256
	expectedDigest := testutil.RandomMultihash(t)
	expectedLocation := testutil.RandomCID(t)
	expectedCause := testutil.RandomCID(t)

	expectedNp := replica.AllocateCaveats{
		Space: expectedSpace,
		Blob: types.Blob{
			Digest: expectedDigest,
			Size:   uint64(expectedSize),
		},
		Site:  expectedLocation,
		Cause: expectedCause,
	}

	node, err := expectedNp.ToIPLD()
	require.NoError(t, err)

	actualNp, err := replica.AllocateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedNp, actualNp)
}

func TestRoundTripAllocateOk(t *testing.T) {
	expectedSize := 256
	expectedTask := testutil.RandomCID(t)

	expectedOk := replica.AllocateOk{
		Size: uint64(expectedSize),
		Site: types.Promise{
			UcanAwait: types.Await{
				Selector: ".out.ok.site",
				Link:     expectedTask,
			},
		},
	}

	node, err := expectedOk.ToIPLD()
	require.NoError(t, err)

	actualNb, err := replica.AllocateOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedOk, actualNb)
}
