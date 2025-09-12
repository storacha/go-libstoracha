package content_test

import (
	"errors"
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/content"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/result/failure"
	fdm "github.com/storacha/go-ucanto/core/result/failure/datamodel"

	"github.com/stretchr/testify/require"
)

func TestRoundTripRetrieveCaveats(t *testing.T) {
	digest := testutil.RandomMultihash(t)
	nb := content.RetrieveCaveats{
		Blob: content.BlobDigest{
			Digest: digest,
		},
		Range: content.Range{
			Start: 123,
			End:   456,
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := content.RetrieveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripRetrieveOk(t *testing.T) {
	ok := content.RetrieveOk{}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := content.RetrieveOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestRoundTripNotFoundError(t *testing.T) {
	expectError := content.NewNotFoundError("blob not found error")

	node, err := expectError.ToIPLD()
	require.NoError(t, err)

	actualError, err := content.NotFoundErrorReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectError, actualError)
}

func TestReadInvalidNotFoundError(t *testing.T) {
	t.Run("missing name", func(t *testing.T) {
		expectError := failure.FromError(errors.New("boom"))

		node, err := expectError.ToIPLD()
		require.NoError(t, err)

		_, err = content.NotFoundErrorReader.Read(node)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("different name", func(t *testing.T) {
		name := "Boom"
		expectError := failure.FromFailureModel(fdm.FailureModel{
			Name:    &name,
			Message: "boom",
		})

		node, err := expectError.ToIPLD()
		require.NoError(t, err)

		_, err = content.NotFoundErrorReader.Read(node)
		require.Error(t, err)
		t.Log(err)
	})
}

func TestRoundTripRangeNotSatisfiableError(t *testing.T) {
	expectError := content.NewRangeNotSatisfiableError("some range not satisfiable error")

	node, err := expectError.ToIPLD()
	require.NoError(t, err)

	actualError, err := content.RangeNotSatisfiableErrorReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, expectError, actualError)
}

func TestReadInvalidRangeNotSatisfiableError(t *testing.T) {
	t.Run("missing name", func(t *testing.T) {
		expectError := failure.FromError(errors.New("boom"))

		node, err := expectError.ToIPLD()
		require.NoError(t, err)

		_, err = content.RangeNotSatisfiableErrorReader.Read(node)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("different name", func(t *testing.T) {
		name := "Boom"
		expectError := failure.FromFailureModel(fdm.FailureModel{
			Name:    &name,
			Message: "boom",
		})

		node, err := expectError.ToIPLD()
		require.NoError(t, err)

		_, err = content.RangeNotSatisfiableErrorReader.Read(node)
		require.Error(t, err)
		t.Log(err)
	})
}
