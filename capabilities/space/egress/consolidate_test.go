package egress_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/egress"
	"github.com/storacha/go-libstoracha/testutil"

	"github.com/stretchr/testify/require"
)

func TestRoundTripConsolidateCaveats(t *testing.T) {
	cause := testutil.RandomCID(t)
	nb := egress.ConsolidateCaveats{
		Cause: cause,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := egress.ConsolidateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestNewConsolidateReceiptReader(t *testing.T) {
	_, err := egress.NewConsolidateReceiptReader()
	require.NoError(t, err)
}

func TestRoundTripConsolidateOk(t *testing.T) {
	errors := []egress.ReceiptError{
		{
			Name:    "SomeReceiptError",
			Message: "some receipt error message",
			Receipt: testutil.RandomCID(t),
		},
		{
			Name:    "SomeReceiptError",
			Message: "some other receipt error message",
			Receipt: testutil.RandomCID(t),
		},
	}
	ok := egress.ConsolidateOk{
		Errors: errors,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := egress.ConsolidateOkReader.Read(node)
	require.NoError(t, err)
	require.ElementsMatch(t, ok.Errors, rok.Errors)
}

func TestRoundTripConsolidateError(t *testing.T) {
	consolidateErr := egress.NewConsolidateError("some egress consolidate error")

	node, err := consolidateErr.ToIPLD()
	require.NoError(t, err)

	readConsolidateErr, err := egress.ConsolidateErrorReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, consolidateErr, readConsolidateErr)
}
