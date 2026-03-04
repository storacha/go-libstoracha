package shard_test

import (
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/storacha/go-libstoracha/capabilities/upload/shard"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/message"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/ran"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/transport/car"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func TestListSerde(t *testing.T) {
	testCases := []struct {
		name string
		nb   shard.ListCaveats
		out  result.Result[shard.ListOk, shard.ListError]
	}{
		{
			name: "defaults",
			nb:   shard.ListCaveats{Root: testutil.RandomCID(t)},
			out: result.Ok[shard.ListOk, shard.ListError](shard.ListOk{
				Size:    1,
				Results: []ipld.Link{testutil.RandomCID(t)},
			}),
		},
		{
			name: "with size",
			nb: shard.ListCaveats{
				Root: testutil.RandomCID(t),
				Size: ptr[uint64](2),
			},
			out: result.Ok[shard.ListOk, shard.ListError](shard.ListOk{
				Cursor:  ptr("cursor1"),
				Size:    3,
				Results: []ipld.Link{testutil.RandomCID(t), testutil.RandomCID(t)},
			}),
		},
		{
			name: "with cursor",
			nb: shard.ListCaveats{
				Root:   testutil.RandomCID(t),
				Cursor: ptr("cursor1"),
			},
			out: result.Ok[shard.ListOk, shard.ListError](shard.ListOk{
				Size:    3,
				Results: []ipld.Link{testutil.RandomCID(t)},
			}),
		},
		{
			name: "with size and cursor",
			nb: shard.ListCaveats{
				Root:   testutil.RandomCID(t),
				Size:   ptr[uint64](2),
				Cursor: ptr("cursor1"),
			},
			out: result.Ok[shard.ListOk, shard.ListError](shard.ListOk{
				Size:    3,
				Results: []ipld.Link{testutil.RandomCID(t)},
			}),
		},
		{
			name: "with error",
			nb: shard.ListCaveats{
				Root: testutil.RandomCID(t),
			},
			out: result.Error[shard.ListOk](shard.NewUploadNotFoundError(testutil.RandomCID(t))),
		},
	}

	for _, tc := range testCases {
		t.Run("round trip "+tc.name, func(t *testing.T) {
			inv, err := shard.List.Invoke(
				testutil.Alice,
				testutil.Bob,
				testutil.Alice.DID().String(),
				tc.nb,
			)
			require.NoError(t, err)

			rcpt := result.MatchResultR1(
				tc.out,
				func(o shard.ListOk) receipt.AnyReceipt {
					rcpt, err := receipt.Issue(
						testutil.Bob,
						result.Ok[shard.ListOk, shard.ListError](o),
						ran.FromInvocation(inv),
					)
					require.NoError(t, err)
					return rcpt
				},
				func(x shard.ListError) receipt.AnyReceipt {
					rcpt, err := receipt.Issue(
						testutil.Bob,
						result.Error[shard.ListOk](x),
						ran.FromInvocation(inv),
					)
					require.NoError(t, err)
					return rcpt
				},
			)

			// round trip the invocation and receipt in an agent message to ensure the
			// invocation can be encoded and the receipt decoded
			msg := roundTripAgentMessage(t, []invocation.Invocation{inv}, []receipt.AnyReceipt{rcpt})

			rcptLink, ok := msg.Get(inv.Link())
			require.True(t, ok)

			reader, err := shard.NewListReceiptReader()
			require.NoError(t, err)

			actualRcpt, err := reader.Read(rcptLink, msg.Blocks())
			require.NoError(t, err)

			// match the expected result with the actual result
			result.MatchResultR0(
				tc.out,
				func(expected shard.ListOk) {
					result.MatchResultR0(
						actualRcpt.Out(),
						func(actual shard.ListOk) {
							require.Equal(t, expected.Cursor, actual.Cursor)
							require.Equal(t, expected.Size, actual.Size)
							require.Equal(t, expected.Results, actual.Results)
						},
						func(actual shard.ListError) {
							require.FailNowf(t, "expected success but got failure", "expected: %v, actual: %v", expected, actual)
						},
					)
				},
				func(expected shard.ListError) {
					result.MatchResultR0(
						actualRcpt.Out(),
						func(actual shard.ListOk) {
							require.FailNowf(t, "expected failure but got success", "expected: %v, actual: %v", expected, actual)
						},
						func(actual shard.ListError) {
							require.Equal(t, expected.ErrorName, actual.ErrorName)
							require.Equal(t, expected.Name(), actual.Name())
							require.Equal(t, expected.Message, actual.Message)
							require.Equal(t, expected.Error(), actual.Error())
						},
					)
				},
			)
		})
	}
}

func TestListDerives(t *testing.T) {
	root := testutil.RandomCID(t)
	testCases := []struct {
		name          string
		claimed       ucan.Capability[shard.ListCaveats]
		delegated     ucan.Capability[shard.ListCaveats]
		expectFailure bool
	}{
		{
			name: "success with defaults",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root},
			),
			expectFailure: false,
		},
		{
			name: "success with cursor",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Cursor: ptr("cursor")},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Cursor: ptr("cursor")},
			),
			expectFailure: false,
		},
		{
			name: "success with size",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Size: ptr[uint64](42)},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Size: ptr[uint64](42)},
			),
			expectFailure: false,
		},
		{
			name: "constraint violation with different root",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: testutil.RandomCID(t)},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root},
			),
			expectFailure: true,
		},
		{
			name: "constraint violation with different cursor",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Cursor: ptr("different-cursor")},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Cursor: ptr("cursor")},
			),
			expectFailure: true,
		},
		{
			name: "constraint escalation with cursor",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Cursor: ptr("cursor")},
			),
			expectFailure: true,
		},
		{
			name: "constraint violation with different size",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Size: ptr[uint64](43)},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Size: ptr[uint64](42)},
			),
			expectFailure: true,
		},
		{
			name: "constraint escalation with size",
			claimed: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root},
			),
			delegated: ucan.NewCapability(
				shard.ListAbility,
				testutil.Alice.DID().String(),
				shard.ListCaveats{Root: root, Size: ptr[uint64](42)},
			),
			expectFailure: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := shard.ListDerive(tc.claimed, tc.delegated)
			if tc.expectFailure {
				t.Log(err)
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func roundTripAgentMessage(t *testing.T, invs []invocation.Invocation, rcpts []receipt.AnyReceipt) message.AgentMessage {
	t.Helper()
	inMsg, err := message.Build(invs, rcpts)
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
