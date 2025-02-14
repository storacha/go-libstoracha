package types

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"

	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-libstoracha/piece/pkg/piece"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/core/schema/options"
	"github.com/storacha/go-ucanto/did"
)

var ErrWrongLength = errors.New("length must be 32")

var MultiaddrConverter = options.NamedBytesConverter("Multiaddr", multiaddr.NewMultiaddrBytes, func(m multiaddr.Multiaddr) ([]byte, error) {
	return m.Bytes(), nil
})

var HasMultihashConverter = options.NamedAnyConverter("HasMultihash", func(nd datamodel.Node) (HasMultihash, error) {
	return linkOrDigest.Read(nd)
}, func(h HasMultihash) (datamodel.Node, error) {
	return h.ToIPLD()
})

var DIDConverter = options.NamedBytesConverter("DID", func(bytes []byte) (did.DID, error) {
	if len(bytes) == 0 {
		return did.Undef, nil
	}
	return did.Decode(bytes)
}, func(did did.DID) ([]byte, error) { return did.Bytes(), nil })

var URLConverter = options.NamedStringConverter("URL",
	func(s string) (url.URL, error) { return schema.URI().Read(s) },
	func(url url.URL) (string, error) { return url.String(), nil })

func headerToNode(h http.Header) (datamodel.Node, error) {
	keys := slices.Collect(maps.Keys(h))
	slices.Sort(keys)

	headers, err := headersToMap(h)
	if err != nil {
		return nil, err
	}
	return ipld.WrapWithRecovery(&HeadersModel{
		Keys:   keys,
		Values: headers,
	}, HeadersType())
}

func headersToMap(h http.Header) (map[string]string, error) {
	headers := map[string]string{}
	for k, v := range h {
		if len(v) > 1 {
			return nil, fmt.Errorf("unsupported multiple values in header: %s", k)
		}
		headers[k] = v[0]
	}
	return headers, nil
}

func nodeToHeader(nd datamodel.Node) (http.Header, error) {
	model, err := ipld.Rebind[HeadersModel](nd, HeadersType())
	if err != nil {
		return nil, err
	}
	header := make(http.Header, len(model.Values))
	for k, v := range model.Values {
		header.Set(k, v)
	}
	return header, nil
}

var HTTPHeaderConverter = options.NamedAnyConverter("HTTPHeaders", nodeToHeader, headerToNode)

var Version1LinkConverter = options.NamedLinkConverter("V1Link", func(c cid.Cid) (ipld.Link, error) {
	return schema.Link(schema.WithVersion(1)).Read(cidlink.Link{Cid: c})
}, func(link ipld.Link) (cid.Cid, error) {
	cl, ok := link.(cidlink.Link)
	if !ok {
		return cid.Undef, errors.New("unsupported link type")
	}
	return cl.Cid, nil
})

var PieceLinkConverter = options.NamedLinkConverter("PieceLink", func(c cid.Cid) (piece.PieceLink, error) {
	return piece.FromLink(cidlink.Link{Cid: c})
}, func(p piece.PieceLink) (cid.Cid, error) {
	return p.Link().(cidlink.Link).Cid, nil
})

var MerkleNodeConverter = options.NamedBytesConverter("Node", func(b []byte) (merkletree.Node, error) {
	if len(b) != len(merkletree.Node{}) {
		return merkletree.Node{}, ErrWrongLength
	}
	return *(*merkletree.Node)(b), nil
}, func(n merkletree.Node) ([]byte, error) {
	return n[:], nil
})

var Converters = []bindnode.Option{MultiaddrConverter, HasMultihashConverter, DIDConverter, URLConverter, HTTPHeaderConverter, Version1LinkConverter, PieceLinkConverter, MerkleNodeConverter}
