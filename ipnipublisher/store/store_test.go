package store_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-varint"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/stretchr/testify/require"
)

var customMetadataID = multicodec.Code(0x3E0000)

type customMetadata struct {
	Data string
}

func (c *customMetadata) ID() multicodec.Code {
	return customMetadataID
}

func (c *customMetadata) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer(varint.ToUvarint(uint64(c.ID())))
	nd := bindnode.Wrap(c, customMetadataPrototype().Type())
	if err := dagcbor.Encode(nd, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *customMetadata) ReadFrom(r io.Reader) (n int64, err error) {
	return readFrom(c, r)
}

func readFrom[T any](val *T, r io.Reader) (int64, error) {
	cr := &countingReader{r: r}
	v, err := varint.ReadUvarint(cr)
	if err != nil {
		return cr.readCount, err
	}
	id := multicodec.Code(v)
	if id != customMetadataID {
		return cr.readCount, fmt.Errorf("transport id does not match %s: %s", customMetadataID, id)
	}

	nb := customMetadataPrototype().NewBuilder()
	err = dagcbor.Decode(nb, cr)
	if err != nil {
		return cr.readCount, err
	}
	nd := nb.Build()
	read := bindnode.Unwrap(nd).(*T)
	*val = *read
	return cr.readCount, nil
}

func (c *customMetadata) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	_, err := readFrom(c, r)
	return err
}

var _ metadata.Protocol = (*customMetadata)(nil)

func customMetadataPrototype() schema.TypedPrototype {
	typeSystem, err := ipld.LoadSchemaBytes([]byte(`
	  type CustomMetadata struct {
		  Data String
		}
	`))
	if err != nil {
		panic(fmt.Errorf("failed to load schema: %w", err))
	}
	return bindnode.Prototype((*customMetadata)(nil), typeSystem.TypeByName("CustomMetadata"))
}

type countingReader struct {
	readCount int64
	r         io.Reader
}

func (c *countingReader) ReadByte() (byte, error) {
	b := []byte{0}
	_, err := c.Read(b)
	return b[0], err
}

func (c *countingReader) Read(b []byte) (n int, err error) {
	read, err := c.r.Read(b)
	c.readCount += int64(read)
	return read, err
}

func TestCustomMetadataContext(t *testing.T) {
	mctx := metadata.Default.WithProtocol(customMetadataID, func() metadata.Protocol {
		return &customMetadata{}
	})

	s := store.FromDatastore(datastore.NewMapDatastore(), store.WithMetadataContext(mctx))

	peerID, err := peer.Decode("12D3KooWLpDkh3ZnFARvrQE1n3Ddb9G66YmfKxp3Z6EYMq1MmcoQ")
	require.NoError(t, err)
	contextID := []byte{1, 2, 3}
	md := mctx.New(&customMetadata{Data: "TEST"})

	err = s.PutMetadataForProviderAndContextID(context.Background(), peerID, contextID, md)
	require.NoError(t, err)

	r, err := s.MetadataForProviderAndContextID(context.Background(), peerID, contextID)
	require.NoError(t, err)
	require.Equal(t, md, r)
}
