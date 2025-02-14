package types

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent/qp"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
)

//go:embed types.ipldsch
var typesSchema []byte

var typesTs = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := ipld.LoadSchemaBytes(typesSchema)
	if err != nil {
		panic(fmt.Errorf("loading types schema: %w", err))
	}
	return ts
}

type HeadersModel struct {
	Keys   []string
	Values map[string]string
}

func HeadersType() ipldschema.Type {
	return typesTs.TypeByName("Headers")
}

func DigestType() ipldschema.Type {
	return typesTs.TypeByName("Digest")
}

type DigestModel struct {
	Digest []byte
}

type HasMultihash interface {
	hasMultihash()
	ToIPLD() (datamodel.Node, error)
	Hash() mh.Multihash
}

type link struct {
	link datamodel.Link
}

func (l link) hasMultihash() {}

func (l link) Hash() mh.Multihash {
	return l.link.(cidlink.Link).Cid.Hash()
}

func (l link) ToIPLD() (datamodel.Node, error) {
	return basicnode.NewLink(l.link), nil
}

func Link(l datamodel.Link) (HasMultihash, failure.Failure) {
	return link{l}, nil
}

type digest mh.Multihash

func (d digest) hasMultihash() {}

func (d digest) Hash() mh.Multihash {
	return mh.Multihash(d)
}

func (d digest) ToIPLD() (datamodel.Node, error) {
	return qp.BuildMap(basicnode.Prototype.Map, 1, func(ma datamodel.MapAssembler) {
		qp.MapEntry(ma, "digest", qp.Bytes(d))
	})
}

func Digest(d DigestModel) (HasMultihash, failure.Failure) {
	return digest(d.Digest), nil
}

func FromHash(mh mh.Multihash) HasMultihash {
	return digest(mh)
}

var linkOrDigest = schema.Or(schema.Mapped(schema.Link(), Link), schema.Mapped(schema.Struct[DigestModel](DigestType(), nil), Digest))
