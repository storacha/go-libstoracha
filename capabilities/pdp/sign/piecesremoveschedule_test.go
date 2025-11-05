package sign_test

import (
	"math/big"
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

func TestPiecesRemoveSchedule(t *testing.T) {
	piecesRemoveScheduleCaveats := sign.PiecesRemoveScheduleCaveats{
		DataSet: testutil.RandomBigInt(t),
		Pieces:  []*big.Int{testutil.RandomBigInt(t), testutil.RandomBigInt(t)},
	}

	dlg, err := delegation.Delegate(
		testutil.Service,
		testutil.Alice,
		[]ucan.Capability[ucan.NoCaveats]{
			ucan.NewCapability("pdp/sign/*", testutil.Service.DID().String(), ucan.NoCaveats{}),
		},
	)
	require.NoError(t, err)

	inv, err := sign.PiecesRemoveSchedule.Invoke(
		testutil.Alice,
		testutil.Bob,
		testutil.Bob.DID().String(),
		piecesRemoveScheduleCaveats,
		delegation.WithProof(delegation.FromDelegation(dlg)),
	)
	require.NoError(t, err)

	piecesRemoveScheduleOk := sign.PiecesRemoveScheduleOk{
		Signature:  testutil.RandomBytes(t, 128),
		V:          testutil.RandomBigInt(t),
		R:          testutil.RandomBytes(t, 32),
		S:          testutil.RandomBytes(t, 32),
		SignedData: testutil.RandomBytes(t, 128),
		Signer:     testutil.RandomBytes(t, 20),
	}

	rcpt, err := receipt.Issue(
		testutil.Service,
		result.Ok[sign.PiecesRemoveScheduleOk, failure.IPLDBuilderFailure](piecesRemoveScheduleOk),
		ran.FromInvocation(inv),
	)
	require.NoError(t, err)

	msg := roundTripAgentMessage(t, inv, rcpt)

	rtInv, ok, err := msg.Invocation(inv.Link())
	require.True(t, ok)
	require.NoError(t, err)

	nb, err := sign.PiecesRemoveScheduleCaveatsReader.Read(rtInv.Capabilities()[0].Nb())
	require.NoError(t, err)

	t.Logf("expected: %+v", piecesRemoveScheduleCaveats)
	t.Logf("actual: %+v", nb)
	require.Equal(t, piecesRemoveScheduleCaveats, nb)

	rtRcpt, ok, err := msg.Receipt(rcpt.Root().Link())
	require.True(t, ok)
	require.NoError(t, err)

	o, x := result.Unwrap(rtRcpt.Out())
	require.Nil(t, x)

	out, err := sign.PiecesRemoveScheduleOkReader.Read(o)
	require.NoError(t, err)

	t.Logf("expected: %+v", piecesRemoveScheduleOk)
	t.Logf("actual: %+v", out)
	require.Equal(t, piecesRemoveScheduleOk, out)
}
