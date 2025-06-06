package filecoin

import (
	"fmt"
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const FilecoinAbility = "filecoin/*"

func ValidateSpaceDID(did string) failure.Failure {
	if !strings.HasPrefix(did, "did:") {
		return schema.NewSchemaError(fmt.Sprintf("invalid DID format: %s", did))
	}

	return nil
}

var Filecoin = validator.NewCapability(
	FilecoinAbility,
	schema.DIDString(),
	schema.Struct[struct{}](nil, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[struct{}]) failure.Failure {
		// Only check if the DID matches
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"resource '%s' doesn't match delegated '%s'",
				claimed.With(), delegated.With(),
			))
		}
		return nil
	},
)

func equalWith(claimed, delegated string) failure.Failure {
	if claimed != delegated {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed, delegated,
		))
	}

	return nil
}

func equalLink(claimed, delegated ipld.Link, fieldName string) failure.Failure {
	if claimed.String() != delegated.String() {
		return schema.NewSchemaError(fmt.Sprintf(
			"%s '%s' doesn't match delegated '%s'",
			fieldName, claimed, delegated,
		))
	}

	return nil
}

func equalPieceLink(claimed, delegated ipld.Link) failure.Failure {
	return equalLink(claimed, delegated, "piece")
}

func validateFilecoinCapability[T any](
	claimed, delegated ucan.Capability[T],
	checkNb func(claimedNb, delegatedNb T) failure.Failure,
) failure.Failure {
	if err := ValidateSpaceDID(claimed.With()); err != nil {
		return err
	}

	if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
		return fail
	}

	if delegated.Can() == FilecoinAbility {
		return nil
	}

	return checkNb(claimed.Nb(), delegated.Nb())
}
