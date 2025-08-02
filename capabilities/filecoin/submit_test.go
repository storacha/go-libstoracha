package filecoin_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestSubmitCaveatsRoundTrip(t *testing.T) {
	nb := filecoin.SubmitCaveats{
		Content: testutil.RandomCID(t),
		Piece:   testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := filecoin.SubmitCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Content.String(), rnb.Content.String())
	require.Equal(t, nb.Piece.String(), rnb.Piece.String())
}

func TestSubmitOkRoundTrip(t *testing.T) {
	ok := filecoin.SubmitOk{
		Piece: testutil.RandomCID(t),
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.SubmitOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
}
