package sign_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/pdp/sign"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/message"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/ran"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/transport/car"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func TestDataSetCreate(t *testing.T) {
	dataSetCreateCaveats := sign.DataSetCreateCaveats{
		DataSet: testutil.RandomBigInt(t),
		Payee:   testutil.RandomBytes(t, 20),
		Metadata: sign.Metadata{
			Keys:   []string{"foo"},
			Values: map[string]string{"foo": "bar"},
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

	inv, err := sign.DataSetCreate.Invoke(
		testutil.Alice,
		testutil.Bob,
		testutil.Bob.DID().String(),
		dataSetCreateCaveats,
		delegation.WithProof(delegation.FromDelegation(dlg)),
	)
	require.NoError(t, err)

	dataSetCreateOk := sign.DataSetCreateOk{
		Signature:  testutil.RandomBytes(t, 128),
		V:          testutil.RandomBigInt(t),
		R:          testutil.RandomBytes(t, 32),
		S:          testutil.RandomBytes(t, 32),
		SignedData: testutil.RandomBytes(t, 128),
		Signer:     testutil.RandomBytes(t, 20),
	}

	rcpt, err := receipt.Issue(
		testutil.Service,
		result.Ok[sign.DataSetCreateOk, failure.IPLDBuilderFailure](dataSetCreateOk),
		ran.FromInvocation(inv),
	)
	require.NoError(t, err)

	msg := roundTripAgentMessage(t, inv, rcpt)

	rtInv, ok, err := msg.Invocation(inv.Link())
	require.True(t, ok)
	require.NoError(t, err)

	nb, err := sign.DataSetCreateCaveatsReader.Read(rtInv.Capabilities()[0].Nb())
	require.NoError(t, err)

	t.Logf("expected: %+v", dataSetCreateCaveats)
	t.Logf("actual: %+v", nb)
	require.Equal(t, dataSetCreateCaveats, nb)

	rtRcpt, ok, err := msg.Receipt(rcpt.Root().Link())
	require.True(t, ok)
	require.NoError(t, err)

	o, x := result.Unwrap(rtRcpt.Out())
	require.Nil(t, x)

	out, err := sign.DataSetCreateOkReader.Read(o)
	require.NoError(t, err)

	t.Logf("expected: %+v", dataSetCreateOk)
	t.Logf("actual: %+v", out)
	require.Equal(t, dataSetCreateOk, out)
}

func roundTripAgentMessage(t *testing.T, inv invocation.Invocation, rcpt receipt.AnyReceipt) message.AgentMessage {
	t.Helper()
	inMsg, err := message.Build([]invocation.Invocation{inv}, []receipt.AnyReceipt{rcpt})
	require.NoError(t, err)

	outCodec := car.NewOutboundCodec()
	req, err := outCodec.Encode(inMsg)
	require.NoError(t, err)

	inCodec, err := car.NewInboundCodec().Accept(req)
	require.NoError(t, err)

	outMsg, err := inCodec.Decoder().Decode(req)
	require.NoError(t, err)

	return outMsg
}
