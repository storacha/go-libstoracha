package http_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/http"
	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/stretchr/testify/require"
)

func TestRoundTripPutCaveats(t *testing.T) {
	blobAllocCID := testutil.RandomCID(t)
	nb := http.PutCaveats{
		Body: http.Body{
			Digest: testutil.RandomMultihash(t),
			Size:   uint64(1024),
		},
		URL: types.Promise{
			UcanAwait: types.Await{
				Selector: ".out.ok.address.url",
				Link:     blobAllocCID,
			},
		},
		Headers: types.Promise{
			UcanAwait: types.Await{
				Selector: ".out.ok.address.headers",
				Link:     blobAllocCID,
			},
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := http.PutCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripPutOk(t *testing.T) {
	ok := http.PutOk{}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := http.PutOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}
