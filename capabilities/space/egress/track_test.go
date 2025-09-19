package egress_test

import (
	"net/url"
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/egress"
	"github.com/storacha/go-libstoracha/testutil"

	"github.com/stretchr/testify/require"
)

func TestRoundTripTrackCaveats(t *testing.T) {
	batchCID := testutil.RandomCID(t)
	endpoint, _ := url.Parse("http://piri.com/receipts/{cid}")
	nb := egress.TrackCaveats{
		Receipts: batchCID,
		Endpoint: endpoint,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := egress.TrackCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Receipts.String(), rnb.Receipts.String())
	require.Equal(t, nb.Endpoint.String(), rnb.Endpoint.String())
}

func TestNewTrackReceiptReader(t *testing.T) {
	_, err := egress.NewTrackReceiptReader()
	require.NoError(t, err)
}

func TestRoundTripTrackError(t *testing.T) {
	trackErr := egress.NewTrackError("some egress track error")

	node, err := trackErr.ToIPLD()
	require.NoError(t, err)

	readTrackErr, err := egress.TrackErrorReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, trackErr, readTrackErr)
}
