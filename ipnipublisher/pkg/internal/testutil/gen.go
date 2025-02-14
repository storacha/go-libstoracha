package testutil

import (
	crand "crypto/rand"
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	mh "github.com/multiformats/go-multihash"
)

func RandomBytes(size int) []byte {
	bytes := make([]byte, size)
	_, _ = crand.Read(bytes)
	return bytes
}

func RandomCID() datamodel.Link {
	bytes := RandomBytes(10)
	c, _ := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}.Sum(bytes)
	return cidlink.Link{Cid: c}
}

func RandomMultihash() mh.Multihash {
	return RandomCID().(cidlink.Link).Hash()
}

func RandomMultihashes(count int) []mh.Multihash {
	if count <= 0 {
		panic(errors.New("count must be greater than 0"))
	}
	mhs := make([]mh.Multihash, 0, count)
	for range count {
		mhs = append(mhs, RandomMultihash())
	}
	return mhs
}
