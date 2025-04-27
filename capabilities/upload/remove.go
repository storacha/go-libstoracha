package upload

import (
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const (
	RemoveAbility = "upload/remove"
)

type RemoveCaveats struct {
	Root cid.Cid `ipld:"root"`
}

func (rc RemoveCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, RemoveCaveatsType(), types.Converters...)
}

var RemoveCaveatsReader = schema.Struct[RemoveCaveats](RemoveCaveatsType(), nil, types.Converters...)

type RemoveOk struct {
	Root   cid.Cid   `ipld:"root"`
	Shards []cid.Cid `ipld:"shards,omitempty"`
}

func (ro RemoveOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, RemoveOkType(), types.Converters...)
}

var RemoveOkReader = schema.Struct[RemoveOk](RemoveOkType(), nil, types.Converters...)

var Remove = validator.NewCapability(
	RemoveAbility,
	schema.DIDString(),
	RemoveCaveatsReader,
	func(claimed, delegated ucan.Capability[RemoveCaveats]) failure.Failure {
		if err := ValidateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
			return fail
		}

		if delegated.Can() == UploadAbility {
			return nil
		}

		if fail := equalRoot(claimed.Nb().Root, delegated.Nb().Root); fail != nil {
			return fail
		}

		return nil
	},
)
