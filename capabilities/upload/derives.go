package upload

import (
	"fmt"
	"strings"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
)

func validateSpaceDID(did string) failure.Failure {
	if !strings.HasPrefix(did, "did:key:") {
		return schema.NewSchemaError(fmt.Sprintf("expected did:key but got %s", did))
	}

	return nil
}

func equalRoot(claimed, delegated ipld.Link) failure.Failure {
	if claimed.String() != delegated.String() {
		return schema.NewSchemaError(fmt.Sprintf(
			"root '%s' doesn't match delegated '%s'",
			claimed, delegated,
		))
	}

	return nil
}

func equalShards(claimed, delegated []ipld.Link) failure.Failure {
	delegatedMap := make(map[string]bool)
	for _, shard := range delegated {
		delegatedMap[shard.String()] = true
	}

	for _, shard := range claimed {
		if !delegatedMap[shard.String()] {
			return schema.NewSchemaError(fmt.Sprintf(
				"shard '%s' not found in delegated shards",
				shard,
			))
		}
	}

	return nil
}

// AddDerive implements the derivation logic for upload/add capabilities
func AddDerive(claimed, delegated any) failure.Failure {
	// Convert to generic capabilities to access basic properties
	claimedCap, ok := claimed.(interface {
		Can() string
		With() string
		Nb() AddCaveats
	})
	if !ok {
		return schema.NewSchemaError("invalid claimed capability type")
	}

	delegatedCap, ok := delegated.(interface {
		Can() string
		With() string
	})
	if !ok {
		return schema.NewSchemaError("invalid delegated capability type")
	}

	if err := validateSpaceDID(claimedCap.With()); err != nil {
		return err
	}

	if claimedCap.With() != delegatedCap.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimedCap.With(), delegatedCap.With(),
		))
	}

	// Allow derivation from upload/* wildcard capability
	if delegatedCap.Can() == UploadAbility {
		return nil
	}

	// If not wildcard, need to validate specific constraints
	delegatedAddCap, ok := delegated.(interface {
		Nb() AddCaveats
	})
	if !ok {
		return schema.NewSchemaError("delegated capability doesn't have AddCaveats")
	}

	if fail := equalRoot(claimedCap.Nb().Root, delegatedAddCap.Nb().Root); fail != nil {
		return fail
	}

	if len(delegatedAddCap.Nb().Shards) > 0 {
		if fail := equalShards(claimedCap.Nb().Shards, delegatedAddCap.Nb().Shards); fail != nil {
			return fail
		}
	}
	return nil
}

// GetDerive implements the derivation logic for upload/get capabilities
func GetDerive(claimed, delegated any) failure.Failure {
	// Convert to generic capabilities to access basic properties
	claimedCap, ok := claimed.(interface {
		Can() string
		With() string
		Nb() GetCaveats
	})
	if !ok {
		return schema.NewSchemaError("invalid claimed capability type")
	}

	delegatedCap, ok := delegated.(interface {
		Can() string
		With() string
	})
	if !ok {
		return schema.NewSchemaError("invalid delegated capability type")
	}

	if err := validateSpaceDID(claimedCap.With()); err != nil {
		return err
	}

	if claimedCap.With() != delegatedCap.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimedCap.With(), delegatedCap.With(),
		))
	}

	// Allow derivation from upload/* wildcard capability
	if delegatedCap.Can() == UploadAbility {
		return nil
	}

	// If not wildcard, need to validate specific constraints
	delegatedGetCap, ok := delegated.(interface {
		Nb() GetCaveats
	})
	if !ok {
		return schema.NewSchemaError("delegated capability doesn't have GetCaveats")
	}

	if fail := equalRoot(claimedCap.Nb().Root, delegatedGetCap.Nb().Root); fail != nil {
		return fail
	}
	return nil
}

// RemoveDerive implements the derivation logic for upload/remove capabilities
func RemoveDerive(claimed, delegated any) failure.Failure {
	// Convert to generic capabilities to access basic properties
	claimedCap, ok := claimed.(interface {
		Can() string
		With() string
		Nb() RemoveCaveats
	})
	if !ok {
		return schema.NewSchemaError("invalid claimed capability type")
	}

	delegatedCap, ok := delegated.(interface {
		Can() string
		With() string
	})
	if !ok {
		return schema.NewSchemaError("invalid delegated capability type")
	}

	if err := validateSpaceDID(claimedCap.With()); err != nil {
		return err
	}

	if claimedCap.With() != delegatedCap.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimedCap.With(), delegatedCap.With(),
		))
	}

	// Allow derivation from upload/* wildcard capability
	if delegatedCap.Can() == UploadAbility {
		return nil
	}

	// If not wildcard, need to validate specific constraints
	delegatedRemoveCap, ok := delegated.(interface {
		Nb() RemoveCaveats
	})
	if !ok {
		return schema.NewSchemaError("delegated capability doesn't have RemoveCaveats")
	}

	if fail := equalRoot(claimedCap.Nb().Root, delegatedRemoveCap.Nb().Root); fail != nil {
		return fail
	}

	return nil
}

// ListDerive implements the derivation logic for upload/list capabilities
func ListDerive(claimed, delegated any) failure.Failure {
	// Convert to generic capabilities to access basic properties
	claimedCap, ok := claimed.(interface {
		Can() string
		With() string
		Nb() ListCaveats
	})
	if !ok {
		return schema.NewSchemaError("invalid claimed capability type")
	}

	delegatedCap, ok := delegated.(interface {
		Can() string
		With() string
	})
	if !ok {
		return schema.NewSchemaError("invalid delegated capability type")
	}

	if err := validateSpaceDID(claimedCap.With()); err != nil {
		return err
	}

	if claimedCap.With() != delegatedCap.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimedCap.With(), delegatedCap.With(),
		))
	}

	// Allow derivation from upload/* wildcard capability
	if delegatedCap.Can() == UploadAbility {
		return nil
	}

	return nil
}
