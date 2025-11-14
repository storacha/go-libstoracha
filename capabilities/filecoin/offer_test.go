package filecoin_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestOfferCaveatsRoundTrip(t *testing.T) {
	t.Run("no pdp", func(t *testing.T) {
		nb := filecoin.OfferCaveats{
			Content: testutil.RandomCID(t),
			Piece:   testutil.RandomCID(t),
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := filecoin.OfferCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb.Content.String(), rnb.Content.String())
		require.Equal(t, nb.Piece.String(), rnb.Piece.String())
	})
	t.Run("with pdp", func(t *testing.T) {
		pdpLink := testutil.RandomCID(t)
		nb := filecoin.OfferCaveats{
			Content: testutil.RandomCID(t),
			Piece:   testutil.RandomCID(t),
			PDP:     &pdpLink,
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := filecoin.OfferCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb.Content.String(), rnb.Content.String())
		require.Equal(t, nb.Piece.String(), rnb.Piece.String())
		require.NotNil(t, rnb.PDP)
		require.Equal(t, (*nb.PDP).String(), (*rnb.PDP).String())
	})
}

func TestOfferOkRoundTrip(t *testing.T) {
	ok := filecoin.OfferOk{
		Piece: testutil.RandomCID(t),
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.OfferOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
}
