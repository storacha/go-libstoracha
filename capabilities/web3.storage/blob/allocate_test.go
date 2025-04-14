package blob_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/web3.storage/blob"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAllocateCaveats(t *testing.T) {
	digest, bytes := testutil.RandomBytes(t, 256)
	nb := blob.AllocateCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: blob.Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
		Cause: testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.AllocateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripAllocateOk(t *testing.T) {
	testURL, err := url.Parse("http://storacha.network")
	require.NoError(t, err)

	ok := blob.AllocateOk{
		Size: 1024,
		Address: &blob.Address{
			URL:       *testURL,
			Headers:   http.Header{"Testhdr": []string{"test"}},
			ExpiresAt: time.Now().UTC().Truncate(time.Second),
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := blob.AllocateOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}
