package egress_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/storacha/go-libstoracha/capabilities/account/egress"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/did"

	"github.com/stretchr/testify/require"
)

func TestRoundTripGetCaveats(t *testing.T) {
	t.Run("marshals and unmarshals correctly", func(t *testing.T) {
		space1 := testutil.RandomDID(t)
		space2 := testutil.RandomDID(t)

		nb := egress.GetCaveats{
			Spaces: []did.DID{space1, space2},
			Period: &egress.Period{
				From: time.UnixMilli(123000).UTC(),
				To:   time.UnixMilli(456000).UTC(),
			},
		}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := egress.GetCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Equal(t, nb.Spaces[0].String(), rnb.Spaces[0].String())
		require.Equal(t, nb.Spaces[1].String(), rnb.Spaces[1].String())
		require.True(t, nb.Period.From.Equal(rnb.Period.From))
		require.True(t, nb.Period.To.Equal(rnb.Period.To))
	})

	t.Run("properties are optional", func(t *testing.T) {
		nb := egress.GetCaveats{}

		node, err := nb.ToIPLD()
		require.NoError(t, err)

		rnb, err := egress.GetCaveatsReader.Read(node)
		require.NoError(t, err)
		require.Nil(t, rnb.Spaces)
		require.Nil(t, rnb.Period)
	})
}

func TestNewGetReceiptReader(t *testing.T) {
	_, err := egress.NewGetReceiptReader()
	require.NoError(t, err)
}

func TestRoundTripGetOk(t *testing.T) {
	space1 := testutil.RandomDID(t)

	ok := egress.GetOk{
		Total: 1000,
		Spaces: egress.SpacesModel{
			Keys: []did.DID{space1},
			Values: map[did.DID]egress.SpaceEgress{
				space1: {
					Total: 500,
					DailyStats: []egress.DailyStats{
						{
							Date:   time.Now().Truncate(time.Second).UTC(),
							Egress: 250,
						},
						{
							Date:   time.Now().Add(-24 * time.Hour).Truncate(time.Second).UTC(),
							Egress: 250,
						},
					},
				},
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	var buf bytes.Buffer

	err = dagcbor.Encode(node, &buf)
	require.NoError(t, err)

	builder := basicnode.Prototype.Any.NewBuilder()
	err = dagcbor.Decode(builder, &buf)
	require.NoError(t, err)

	rok, err := egress.GetOkReader.Read(builder.Build())
	require.NoError(t, err)

	require.Equal(t, ok.Total, rok.Total)
	require.Equal(t, len(ok.Spaces.Keys), len(rok.Spaces.Keys))
	require.Equal(t, ok.Spaces.Values[space1].Total, rok.Spaces.Values[space1].Total)
	require.Equal(t, len(ok.Spaces.Values[space1].DailyStats), len(rok.Spaces.Values[space1].DailyStats))
	require.True(t, ok.Spaces.Values[space1].DailyStats[0].Date.Equal(rok.Spaces.Values[space1].DailyStats[0].Date))
}
