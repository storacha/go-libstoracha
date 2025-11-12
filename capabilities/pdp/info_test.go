package pdp_test

import (
	"testing"

	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/storacha/go-libstoracha/capabilities/pdp"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripInfoCaveats(t *testing.T) {
	bytes := testutil.RandomBytes(t, 256)
	digest := testutil.MultihashFromBytes(t, bytes)

	ic := pdp.InfoCaveats{
		Blob: digest,
	}

	node, err := ic.ToIPLD()
	require.NoError(t, err)

	ric, err := pdp.InfoCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ic, ric)
}

func TestRoundTripInfoOk(t *testing.T) {
	piece := testutil.RandomPiece(t, 1024)

	aggregate := testutil.RandomPiece(t, 2048)

	// Create some test proof nodes
	node1 := merkletree.ZeroCommitmentForLevel(0)
	node2 := merkletree.ZeroCommitmentForLevel(1)

	inclusionProof := merkletree.ProofData{
		Path:  []merkletree.Node{node1, node2},
		Index: 42,
	}

	io := pdp.InfoOk{
		Piece: piece,
		Aggregates: []pdp.InfoAcceptedAggregate{
			{
				Aggregate:      aggregate,
				InclusionProof: inclusionProof,
			},
		},
	}

	node, err := io.ToIPLD()
	require.NoError(t, err)

	rio, err := pdp.InfoOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, io.Piece, rio.Piece)
	require.Equal(t, len(io.Aggregates), len(rio.Aggregates))
	require.Equal(t, io.Aggregates[0].Aggregate, rio.Aggregates[0].Aggregate)
	require.Equal(t, io.Aggregates[0].InclusionProof, rio.Aggregates[0].InclusionProof)
}
