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
