package testutil

import (
	crand "crypto/rand"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

func RandomCID(t *testing.T) ipld.Link {
	digest, _ := RandomBytes(t, 10)
	return cidlink.Link{Cid: cid.NewCidV1(cid.Raw, digest)}
}

func RandomBytes(t *testing.T, size int) (multihash.Multihash, []byte) {
	bytes := make([]byte, size)
	_, err := crand.Read(bytes)
	require.NoError(t, err)
	digest, err := multihash.Sum(bytes, multihash.SHA2_256, -1)
	require.NoError(t, err)
	return digest, bytes
}
