package filecoin_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestInfoCaveatsRoundTrip(t *testing.T) {
	nb := filecoin.InfoCaveats{
		Piece: testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := filecoin.InfoCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Piece.String(), rnb.Piece.String())
}

func TestInfoOkRoundTrip(t *testing.T) {
	aggregateLink := testutil.RandomCID(t)

	ok := filecoin.InfoOk{
		Piece: testutil.RandomCID(t),
		Aggregates: []filecoin.InfoAcceptedAggregate{
			{
				Aggregate: aggregateLink,
				Inclusion: filecoin.InclusionProof{
					Subtree: []byte{1, 2, 3},
					Index:   []byte{4, 5, 6},
				},
			},
		},
		Deals: []filecoin.InfoAcceptedDeal{
			{
				Aggregate: aggregateLink,
				Aux: filecoin.DealMetadata{
					DataType: 1,
					DataSource: filecoin.SingletonMarketSource{
						DealID: 12345,
					},
				},
				Provider: "f01234",
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.InfoOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
	require.Len(t, rok.Aggregates, 1)
	require.Equal(t, ok.Aggregates[0].Aggregate.String(), rok.Aggregates[0].Aggregate.String())
	require.Len(t, rok.Deals, 1)
	require.Equal(t, ok.Deals[0].Provider, rok.Deals[0].Provider)
}
