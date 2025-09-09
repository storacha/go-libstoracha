package replica_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/storacha/go-libstoracha/capabilities/blob/replica"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/testutil"
)

func TestRoundTripTransferCaveats(t *testing.T) {
	expectedSize := 256
	expectedDigest := testutil.RandomMultihash(t)
	expectedLocation := testutil.RandomCID(t)
	expectedCause := testutil.RandomCID(t)

	expectedNp := replica.TransferCaveats{
		Blob: types.Blob{
			Digest: expectedDigest,
			Size:   uint64(expectedSize),
		},
		Site:  expectedLocation,
		Cause: expectedCause,
	}

	node, err := expectedNp.ToIPLD()
	require.NoError(t, err)

	actualNp, err := replica.TransferCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectedNp, actualNp)
}

func TestRoundTripTransfeOk(t *testing.T) {
	t.Run("with PDP link", func(t *testing.T) {
		expectedLocation := testutil.RandomCID(t)
		expectedPDP := testutil.RandomCID(t)

		expectedOk := replica.TransferOk{
			Site: expectedLocation,
			PDP:  &expectedPDP,
		}

		node, err := expectedOk.ToIPLD()
		require.NoError(t, err)

		actualOk, err := replica.TransferOkReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, expectedOk, actualOk)
	})

	t.Run("without PDP link", func(t *testing.T) {
		expectedLocation := testutil.RandomCID(t)

		expectedOk := replica.TransferOk{
			Site: expectedLocation,
		}

		node, err := expectedOk.ToIPLD()
		require.NoError(t, err)

		actualOk, err := replica.TransferOkReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, expectedOk, actualOk)
	})
}

func TestRoundTripTransferError(t *testing.T) {
	expectError := replica.NewTransferError("some error")

	node, err := expectError.ToIPLD()
	require.NoError(t, err)

	actualError, err := replica.TransferErrorReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectError, actualError)
}
