package upload

import (
	"fmt"
	"strings"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const (
	UploadAbility = "upload/*"
)

func validateSpaceDID(did string) failure.Failure {
	if !strings.HasPrefix(did, "did:") {
		return schema.NewSchemaError(fmt.Sprintf("expected did:key but got %s", did))
	}
	return nil
}

var Upload = validator.NewCapability(
	UploadAbility,
	schema.DIDString(),
	schema.Struct[struct{}](nil, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[struct{}]) failure.Failure {
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

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
