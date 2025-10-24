package index_test

import (
	"bytes"
	"testing"

	"github.com/ipld/go-ipld-prime/codec/dagjson"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/storacha/go-libstoracha/capabilities/space/index"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAddCaveats(t *testing.T) {
	testCases := []struct {
		name string
		nb   index.AddCaveats
	}{
		{
			name: "without content link",
			nb: index.AddCaveats{
				Index: testutil.RandomCID(t),
			},
		},
		{
			name: "with content link",
			nb: index.AddCaveats{
				Index:   testutil.RandomCID(t),
				Content: testutil.RandomCID(t),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := tc.nb.ToIPLD()
			require.NoError(t, err)

			var buf bytes.Buffer

			err = dagjson.Encode(node, &buf)
			require.NoError(t, err)

			t.Log(buf.String())

			builder := basicnode.Prototype.Any.NewBuilder()
			err = dagjson.Decode(builder, &buf)
			require.NoError(t, err)

			rnb, err := index.AddCaveatsReader.Read(builder.Build())
			require.NoError(t, err)
			require.Equal(t, tc.nb, rnb)
		})
	}
}
