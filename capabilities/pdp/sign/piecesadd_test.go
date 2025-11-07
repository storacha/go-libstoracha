package sign_test

import (
	"math/rand/v2"
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/pdp/sign"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/ran"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func TestPiecesAdd(t *testing.T) {
	piecesAddCaveats := sign.PiecesAddCaveats{
		DataSet:    testutil.RandomBigInt(t),
		FirstAdded: testutil.RandomBigInt(t),
		PieceData:  [][]byte{testutil.RandomBytes(t, 32), testutil.RandomBytes(t, 32)},
		Metadata: []sign.Metadata{
			{
				Keys:   []string{"foo"},
				Values: map[string]string{"foo": "bar"},
			},
		},
	}

	dlg, err := delegation.Delegate(
		testutil.Service,
		testutil.Alice,
		[]ucan.Capability[ucan.NoCaveats]{
			ucan.NewCapability("pdp/sign/*", testutil.Service.DID().String(), ucan.NoCaveats{}),
		},
	)
	require.NoError(t, err)

	inv, err := sign.PiecesAdd.Invoke(
		testutil.Alice,
		testutil.Bob,
		testutil.Bob.DID().String(),
		piecesAddCaveats,
		delegation.WithProof(delegation.FromDelegation(dlg)),
	)
	require.NoError(t, err)

	piecesAddOk := sign.PiecesAddOk{
		Signature:  testutil.RandomBytes(t, 128),
		V:          uint8(rand.IntN(255)),
		R:          randomHash(t),
		S:          randomHash(t),
		SignedData: testutil.RandomBytes(t, 128),
		Signer:     randomAddress(t),
	}

	rcpt, err := receipt.Issue(
		testutil.Service,
		result.Ok[sign.PiecesAddOk, failure.IPLDBuilderFailure](piecesAddOk),
		ran.FromInvocation(inv),
	)
	require.NoError(t, err)

	msg := roundTripAgentMessage(t, inv, rcpt)

	rtInv, ok, err := msg.Invocation(inv.Link())
	require.True(t, ok)
	require.NoError(t, err)

	nb, err := sign.PiecesAddCaveatsReader.Read(rtInv.Capabilities()[0].Nb())
	require.NoError(t, err)

	t.Logf("expected: %+v", piecesAddCaveats)
	t.Logf("actual: %+v", nb)
	require.Equal(t, piecesAddCaveats, nb)

	rtRcpt, ok, err := msg.Receipt(rcpt.Root().Link())
	require.True(t, ok)
	require.NoError(t, err)

	o, x := result.Unwrap(rtRcpt.Out())
	require.Nil(t, x)

	out, err := sign.PiecesAddOkReader.Read(o)
	require.NoError(t, err)

	t.Logf("expected: %+v", piecesAddOk)
	t.Logf("actual: %+v", out)
	require.Equal(t, piecesAddOk, out)
}
