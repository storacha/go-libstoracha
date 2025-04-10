package blob_test

import (
	"testing"
	"time"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"

	"github.com/stretchr/testify/require"
)

func TestRoundTripGetCaveats(t *testing.T) {
	digest, _ := testutil.RandomBytes(t, 256)
	nb := blob.GetCaveats{
		Digest: digest,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.GetCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripGetOk(t *testing.T) {
	digest, _ := testutil.RandomBytes(t, 256)
	ok := blob.GetOk{
		Blob: types.Blob{
			Digest: digest,
			Size:   uint64(1024),
		},
		Cause:      testutil.RandomCID(t),
		InsertedAt: time.Now().UTC().Truncate(time.Second),
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := blob.GetOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}
