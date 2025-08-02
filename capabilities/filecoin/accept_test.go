package filecoin_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestAcceptCaveatsRoundTrip(t *testing.T) {

	nb := filecoin.AcceptCaveats{
		Content: testutil.RandomCID(t),
		Piece:   testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := filecoin.AcceptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Content.String(), rnb.Content.String())
	require.Equal(t, nb.Piece.String(), rnb.Piece.String())
}

func TestAcceptOkRoundTrip(t *testing.T) {
	ok := filecoin.AcceptOk{
		Piece:     testutil.RandomCID(t),
		Aggregate: testutil.RandomCID(t),
		Inclusion: filecoin.InclusionProof{
			Subtree: []byte{1, 2, 3},
			Index:   []byte{4, 5, 6},
		},
		Aux: filecoin.DealMetadata{
			DataType: 1,
			DataSource: filecoin.SingletonMarketSource{
				DealID: 12345,
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.AcceptOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
	require.Equal(t, ok.Aggregate.String(), rok.Aggregate.String())
	require.Equal(t, ok.Aux.DataType, rok.Aux.DataType)
	require.Equal(t, ok.Aux.DataSource.DealID, rok.Aux.DataSource.DealID)
}
