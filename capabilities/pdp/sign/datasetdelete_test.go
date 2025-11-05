package sign_test

import (
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

func TestDataSetDelete(t *testing.T) {
	dataSetDeleteCaveats := sign.DataSetDeleteCaveats{
		DataSet: testutil.RandomBigInt(t),
	}

	dlg, err := delegation.Delegate(
		testutil.Service,
		testutil.Alice,
		[]ucan.Capability[ucan.NoCaveats]{
			ucan.NewCapability("pdp/sign/*", testutil.Service.DID().String(), ucan.NoCaveats{}),
		},
	)
	require.NoError(t, err)

	inv, err := sign.DataSetDelete.Invoke(
		testutil.Alice,
		testutil.Bob,
		testutil.Bob.DID().String(),
		dataSetDeleteCaveats,
		delegation.WithProof(delegation.FromDelegation(dlg)),
	)
	require.NoError(t, err)

	dataSetDeleteOk := sign.DataSetDeleteOk{
		Signature:  testutil.RandomBytes(t, 128),
		V:          testutil.RandomBigInt(t),
		R:          testutil.RandomBytes(t, 32),
		S:          testutil.RandomBytes(t, 32),
		SignedData: testutil.RandomBytes(t, 128),
		Signer:     testutil.RandomBytes(t, 20),
	}

	rcpt, err := receipt.Issue(
		testutil.Service,
		result.Ok[sign.DataSetDeleteOk, failure.IPLDBuilderFailure](dataSetDeleteOk),
		ran.FromInvocation(inv),
	)
	require.NoError(t, err)

	msg := roundTripAgentMessage(t, inv, rcpt)

	rtInv, ok, err := msg.Invocation(inv.Link())
	require.True(t, ok)
	require.NoError(t, err)

	nb, err := sign.DataSetDeleteCaveatsReader.Read(rtInv.Capabilities()[0].Nb())
	require.NoError(t, err)

	t.Logf("expected: %+v", dataSetDeleteCaveats)
	t.Logf("actual: %+v", nb)
	require.Equal(t, dataSetDeleteCaveats, nb)

	rtRcpt, ok, err := msg.Receipt(rcpt.Root().Link())
	require.True(t, ok)
	require.NoError(t, err)

	o, x := result.Unwrap(rtRcpt.Out())
	require.Nil(t, x)

	out, err := sign.DataSetDeleteOkReader.Read(o)
	require.NoError(t, err)

	t.Logf("expected: %+v", dataSetDeleteOk)
	t.Logf("actual: %+v", out)
	require.Equal(t, dataSetDeleteOk, out)
}
