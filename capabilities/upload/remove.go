package upload

import (
	"fmt"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const RemoveAbility = "upload/remove"

type RemoveCaveats struct {
	Root ipld.Link
}

func (rc RemoveCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&rc, RemoveCaveatsType(), types.Converters...)
}

var RemoveCaveatsReader = schema.Struct[RemoveCaveats](RemoveCaveatsType(), nil, types.Converters...)

type RemoveOk struct {
	Root   ipld.Link
	Shards []ipld.Link
}

func (ro RemoveOk) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&ro, RemoveOkType(), types.Converters...)
}

var RemoveOkReader = schema.Struct[RemoveOk](RemoveOkType(), nil, types.Converters...)

var Remove = validator.NewCapability(
	RemoveAbility,
	schema.DIDString(),
	RemoveCaveatsReader,
	func(claimed, delegated ucan.Capability[RemoveCaveats]) failure.Failure {
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
			return fail
		}

		if claimed.Nb().Root.String() != delegated.Nb().Root.String() {
			return schema.NewSchemaError(fmt.Sprintf(
				"root '%s' doesn't match delegated '%s'",
				claimed.Nb().Root, delegated.Nb().Root,
			))
		}

		return nil
	},
)
