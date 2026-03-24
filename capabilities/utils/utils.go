package utils

import (
	"fmt"

	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
)

// EqualWith validates that the claimed capability's `with` field matches the delegated one.
func EqualWith[Caveats any](claimed, delegated ucan.Capability[Caveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}
	return nil
}
