package content_test

import (
	"errors"
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/content"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/result/failure"
	fdm "github.com/storacha/go-ucanto/core/result/failure/datamodel"
	"github.com/storacha/go-ucanto/ucan"

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

// TestRetrieveDerive tests the RetrieveDerive function in isolation
func TestRetrieveDerive(t *testing.T) {
	digest := testutil.RandomMultihash(t)
	spaceDID := "did:example:space"

	delegatedCaveats := content.RetrieveCaveats{
		Blob: content.BlobDigest{Digest: digest},
		Range: content.Range{
			Start: 0,
			End:   1024,
		},
	}

	delegated := ucan.NewCapability(
		content.RetrieveAbility,
		spaceDID,
		delegatedCaveats,
	)

	t.Run("accepts an identical capability", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			delegated.Nb(),
		)

		fail := content.RetrieveDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("rejects the wrong space", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			"did:example:different-space",
			delegated.Nb(),
		)

		fail := content.RetrieveDerive(claimed, delegated)
		require.ErrorContains(t, fail, "Resource 'did:example:different-space' doesn't match delegated 'did:example:space'")
	})

	t.Run("rejects a different blob digest", func(t *testing.T) {
		differentDigest := testutil.RandomMultihash(t)
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			content.RetrieveCaveats{
				Blob:  content.BlobDigest{Digest: differentDigest},
				Range: delegated.Nb().Range,
			},
		)

		fail := content.RetrieveDerive(claimed, delegated)
		require.Error(t, fail)
		require.ErrorContains(t, fail, "Digest")
	})

	t.Run("accepts a subset range", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			content.RetrieveCaveats{
				Blob: delegated.Nb().Blob,
				Range: content.Range{
					Start: 100,
					End:   500,
				},
			},
		)

		fail := content.RetrieveDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("rejects range exceeding upper bound", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			content.RetrieveCaveats{
				Blob: delegated.Nb().Blob,
				Range: content.Range{
					Start: 512,
					End:   2048, // Exceeds delegated end of 1024
				},
			},
		)

		fail := content.RetrieveDerive(claimed, delegated)
		require.ErrorContains(t, fail, "End offset 2048 violates imposed 1024 constraint")
	})

	t.Run("rejects range below lower bound", func(t *testing.T) {
		delegatedWithOffset := ucan.NewCapability(
			content.RetrieveAbility,
			spaceDID,
			content.RetrieveCaveats{
				Blob: content.BlobDigest{Digest: digest},
				Range: content.Range{
					Start: 100,
					End:   1024,
				},
			},
		)

		claimed := ucan.NewCapability(
			delegatedWithOffset.Can(),
			delegatedWithOffset.With(),
			content.RetrieveCaveats{
				Blob: delegatedWithOffset.Nb().Blob,
				Range: content.Range{
					Start: 0, // Starts before delegated start of 100
					End:   500,
				},
			},
		)

		fail := content.RetrieveDerive(claimed, delegatedWithOffset)
		require.ErrorContains(t, fail, "Start offset 0 violates imposed 100 constraint")
	})

	t.Run("accepts range equal to bounds", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			content.RetrieveCaveats{
				Blob: delegated.Nb().Blob,
				Range: content.Range{
					Start: 0,
					End:   1024,
				},
			},
		)

		fail := content.RetrieveDerive(claimed, delegated)
		require.NoError(t, fail)
	})
}
