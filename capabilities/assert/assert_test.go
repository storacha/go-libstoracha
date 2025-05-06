package assert_test

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"

	"github.com/ipld/go-ipld-prime/codec/dagjson"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripLocationCaveats(t *testing.T) {
	digest, _ := testutil.RandomBytes(t, 256)
	location, err := url.Parse("http://localhost")
	require.NoError(t, err)

	length := uint64(20)
	nb := assert.LocationCaveats{
		Content:  types.FromHash(digest),
		Space:    testutil.RandomPrincipal(t).DID(),
		Location: []url.URL{*location},
		Range: &assert.Range{
			Offset: 100,
			Length: &length,
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	buf := bytes.NewBuffer([]byte{})
	err = dagjson.Encode(node, buf)
	require.NoError(t, err)

	fmt.Println(buf.String())

	builder := basicnode.Prototype.Any.NewBuilder()
	err = dagjson.Decode(builder, buf)
	require.NoError(t, err)

	rnb, err := assert.LocationCaveatsReader.Read(builder.Build())
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}
