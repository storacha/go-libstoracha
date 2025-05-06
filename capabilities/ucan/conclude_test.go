package ucan_test

import (
	"testing"
	"time"

	"github.com/storacha/go-libstoracha/capabilities/ucan"
	"github.com/storacha/go-libstoracha/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripConcludeCaveats(t *testing.T) {
	nb := ucan.ConcludeCaveats{
		Receipt: testutil.RandomCID(t),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := ucan.ConcludeCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripConcludeOk(t *testing.T) {
	ok := ucan.ConcludeOk{
		Time: time.Now().Truncate(time.Millisecond),
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := ucan.ConcludeOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}
